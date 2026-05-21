package server

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const discordCategoryName = "RealGuides"

type DiscordSync struct {
	db      *DB
	games   *GameInfo
	session *discordgo.Session
}

func NewDiscordSync(db *DB, games *GameInfo) (*DiscordSync, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_BOT_TOKEN not set")
	}
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}
	return &DiscordSync{db: db, games: games, session: sess}, nil
}

// SyncPeriodically runs an immediate sync then repeats on the given interval.
func (ds *DiscordSync) SyncPeriodically(interval time.Duration) {
	if err := ds.Sync(); err != nil {
		log.Printf("discord: initial sync: %v", err)
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for range t.C {
		if err := ds.Sync(); err != nil {
			log.Printf("discord: sync: %v", err)
		}
	}
}

// Sync fetches reviews from all guilds the bot belongs to.
func (ds *DiscordSync) Sync() error {
	guilds, err := ds.session.UserGuilds(100, "", "", false)
	if err != nil {
		return fmt.Errorf("fetch guilds: %w", err)
	}
	for _, g := range guilds {
		if err := ds.syncGuild(g.ID); err != nil {
			log.Printf("discord: guild %s (%s): %v", g.Name, g.ID, err)
		}
	}
	return nil
}

func (ds *DiscordSync) syncGuild(guildID string) error {
	channels, err := ds.session.GuildChannels(guildID)
	if err != nil {
		return fmt.Errorf("fetch channels: %w", err)
	}

	var categoryID string
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildCategory &&
			strings.EqualFold(ch.Name, discordCategoryName) {
			categoryID = ch.ID
			break
		}
	}
	if categoryID == "" {
		return nil
	}

	for _, ch := range channels {
		if ch.ParentID != categoryID {
			continue
		}
		if ch.Type != discordgo.ChannelTypeGuildText &&
			ch.Type != discordgo.ChannelTypeGuildForum {
			continue
		}
		gameID := strings.ToLower(ch.Name)
		if err := ds.syncChannel(ch.ID, gameID); err != nil {
			log.Printf("discord: channel %s (game=%s): %v", ch.Name, gameID, err)
		}
	}
	return nil
}

func (ds *DiscordSync) syncChannel(channelID, gameID string) error {
	msgs, err := ds.session.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		return fmt.Errorf("fetch messages: %w", err)
	}
	for _, msg := range msgs {
		review, ok := parseDiscordReview(msg.Content, msg.Author.Username, msg.Author.ID, gameID)
		if !ok {
			continue
		}
		if err := ds.db.UpsertDiscordReview(review, msg.ID); err != nil {
			log.Printf("discord: upsert msg %s: %v", msg.ID, err)
		}
	}
	return nil
}

// ratingRe matches a bullet line like " - Graphics: 10"
var ratingRe = regexp.MustCompile(`(?i)^\s*-\s*(\w+)\s*:\s*(\d+)`)

// parseDiscordReview parses the following post format (username comes from Discord):
//
//	Rating
//	 - Overall: 9
//	 - Graphics: 10
//	 - LearningCurve: 8
//	 - Sound: 7
//	 - LastingAppeal: 9
//	Version: Full
//	Location: Texas
//	Comments: Review text here.
func parseDiscordReview(content, username, userID, gameID string) (ReviewData, bool) {
	r := ReviewData{
		GameID:   gameID,
		UserID:   userID,
		Username: username,
		Reviewer: username,
	}

	inRating := false
	var commentLines []string

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		switch {
		case strings.EqualFold(trimmed, "rating"):
			inRating = true

		case strings.HasPrefix(lower, "comments:"):
			inRating = false
			if rest := strings.TrimSpace(trimmed[len("comments:"):]); rest != "" {
				commentLines = append(commentLines, rest)
			}

		case strings.HasPrefix(lower, "version:"):
			r.Version = strings.TrimSpace(trimmed[len("version:"):])

		case strings.HasPrefix(lower, "location:"):
			r.Location = strings.TrimSpace(trimmed[len("location:"):])

		case inRating:
			if m := ratingRe.FindStringSubmatch(line); m != nil {
				val, _ := strconv.Atoi(m[2])
				switch strings.ToLower(m[1]) {
				case "overall":
					r.RatingOverall = val
				case "graphics":
					r.RatingGraphics = val
				case "learningcurve", "learning_curve":
					r.RatingLearningCurve = val
				case "sound":
					r.RatingSound = val
				case "lastingappeal", "lasting_appeal":
					r.RatingLastingAppeal = val
				}
			}

		case len(commentLines) > 0:
			commentLines = append(commentLines, line)
		}
	}

	r.Comments = strings.TrimSpace(strings.Join(commentLines, "\n"))

	total := r.RatingOverall + r.RatingGraphics + r.RatingSound +
		r.RatingLearningCurve + r.RatingLastingAppeal
	if total == 0 {
		return ReviewData{}, false
	}
	return r, true
}
