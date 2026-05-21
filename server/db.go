package server

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", path, err)
	}
	db := &DB{conn: conn}
	return db, db.migrate()
}

func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS reviews (
			id                    INTEGER PRIMARY KEY AUTOINCREMENT,
			game_id               TEXT    NOT NULL,
			user_id               TEXT    NOT NULL,
			username              TEXT    NOT NULL DEFAULT '',
			reviewer              TEXT    NOT NULL DEFAULT '',
			location              TEXT    NOT NULL DEFAULT '',
			favorite_games        TEXT    NOT NULL DEFAULT '',
			version               TEXT    NOT NULL DEFAULT '',
			rating_overall        INTEGER NOT NULL DEFAULT 0,
			rating_graphics       INTEGER NOT NULL DEFAULT 0,
			rating_learning_curve INTEGER NOT NULL DEFAULT 0,
			rating_sound          INTEGER NOT NULL DEFAULT 0,
			rating_lasting_appeal INTEGER NOT NULL DEFAULT 0,
			comments              TEXT    NOT NULL DEFAULT '',
			source                TEXT    NOT NULL DEFAULT 'local',
			discord_message_id    TEXT    UNIQUE,
			created_at            DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE UNIQUE INDEX IF NOT EXISTS uq_local_review
			ON reviews (game_id, user_id) WHERE source = 'local';
	`)
	return err
}

func (db *DB) UpsertDiscordReview(r ReviewData, msgID string) error {
	_, err := db.conn.Exec(`
		INSERT INTO reviews (
			game_id, user_id, username, reviewer, location, favorite_games, version,
			rating_overall, rating_graphics, rating_learning_curve, rating_sound,
			rating_lasting_appeal, comments, source, discord_message_id
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,'discord',?)
		ON CONFLICT(discord_message_id) DO UPDATE SET
			rating_overall        = excluded.rating_overall,
			rating_graphics       = excluded.rating_graphics,
			rating_learning_curve = excluded.rating_learning_curve,
			rating_sound          = excluded.rating_sound,
			rating_lasting_appeal = excluded.rating_lasting_appeal,
			comments              = excluded.comments
	`,
		r.GameID, r.UserID, r.Username, r.Reviewer, r.Location, r.FavoriteGames, r.Version,
		r.RatingOverall, r.RatingGraphics, r.RatingLearningCurve, r.RatingSound,
		r.RatingLastingAppeal, r.Comments, msgID,
	)
	return err
}

func (db *DB) InsertLocalReview(r ReviewData) error {
	_, err := db.conn.Exec(`
		INSERT OR IGNORE INTO reviews (
			game_id, user_id, username, reviewer, location, favorite_games, version,
			rating_overall, rating_graphics, rating_learning_curve, rating_sound,
			rating_lasting_appeal, comments, source
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,'local')
	`,
		r.GameID, r.UserID, r.Username, r.Reviewer, r.Location, r.FavoriteGames, r.Version,
		r.RatingOverall, r.RatingGraphics, r.RatingLearningCurve, r.RatingSound,
		r.RatingLastingAppeal, r.Comments,
	)
	return err
}

func (db *DB) GetReviewsByGameID(gameID string) ([]ReviewData, error) {
	rows, err := db.conn.Query(`
		SELECT user_id, username, reviewer, location, favorite_games, version,
		       rating_overall, rating_graphics, rating_learning_curve, rating_sound,
		       rating_lasting_appeal, comments
		FROM reviews
		WHERE game_id = ?
		ORDER BY created_at ASC
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []ReviewData
	for rows.Next() {
		r := ReviewData{GameID: gameID}
		if err := rows.Scan(
			&r.UserID, &r.Username, &r.Reviewer, &r.Location, &r.FavoriteGames, &r.Version,
			&r.RatingOverall, &r.RatingGraphics, &r.RatingLearningCurve, &r.RatingSound,
			&r.RatingLastingAppeal, &r.Comments,
		); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

func (db *DB) Close() error {
	return db.conn.Close()
}
