defaultStatus = "";
pageid="gamesforum";

function rgxa(href,areaid,type,dest)
{
	P_showPop=false;
	document.location.href="http://"+href;
}

function rgia(href,areaid,type,dest)
{
	P_showPop=false;
	document.location.href="http://realguide.real.com/"+href;
}

function doPopup(url,width,height)
{
	newwindow = window.open(url,"realguidepopup","width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=1,menubar=1,scrollbars=1,resizable=1");
	newwindow.focus();
	return newwindow;
}

function doPopupFalse(url,width,height)
{
	newwindow = window.open(url,"realguidepopup","width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=0,menubar=0,scrollbars=0,resizable=0");
	newwindow.focus();
}

function cmpPopup(url,width,height)
{
	newwindow = window.open(url,"realguidepopup","width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=0,menubar=0,scrollbars=0,resizable=0,top=60,left=60");
	newwindow.focus();
}

function cmpPopupNamed(url,name,width,height)
{
	newwindow = window.open(url,name,"width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=0,menubar=0,scrollbars=0,resizable=0,top=60,left=60");
	newwindow.focus();
}

function doNamedPopup(name,url,width,height)
{
	newwindow = window.open(url,name,"width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=1,menubar=1,scrollbars=1,resizable=1");
	newwindow.focus();
	return newwindow;
}

function fnSet(){ 
	oHomePage.setHomePage("http://realguide.real.com");
	event.returnValue = false;
}

function getArray(ArrName)
{
	document.sta_find.s.length = 0;
	for (i = 0; i < ArrName.length; i++)
	{
		if (i == 0)
		{
			document.sta_find.s.options[i] = new Option( ArrName[i], '' );
		}
		else
		{
			document.sta_find.s.options[i] = new Option( ArrName[i], ArrName[i] );
		}
	}
	document.sta_find.s.selectedIndex = 0;
}

function dualProx(mediaurl,pageurl,type)
{
	if( type == "embedded" )
		var newWin=window.open(mediaurl,newWin);
	else
		document.location.href= mediaurl;

	timerID = setTimeout("secondSpawn('"+pageurl+"')",5000);
	timerRunning = true;
}

function secondSpawn(url) 
{
	window.self.location=url;
}
