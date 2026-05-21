defaultStatus = "";

function rgxa(href,areaid,type,dest)
{
	P_showPop=false;
	document.location.href="/RGX/"+pageid+"."+areaid+"."+type+"."+dest+".RGX/"+href;
}

function rgia(href,areaid,type,dest)
{
	P_showPop=false;
	document.location.href="/RGI/"+pageid+"."+areaid+"."+type+"."+dest+".RGI/"+href;
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

function cmpPopupNamedOptions(url,name,width,height,options)
{
	newwindow = window.open(url,name,"width="+width+",height="+height+","+options+",top=60,left=60");
	newwindow.focus();
}

function rezPopupNamed(url,name,width,height)
{
	newwindow = window.open(url,name,"width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=0,menubar=0,scrollbars=0,resizable=1,top=60,left=60");
	newwindow.focus();
}

function scrollNamedPopup(name,url,width,height)
{
	newwindow = window.open(url,name,"width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=0,menubar=0,scrollbars=1,resizable=0,top=60,left=640");
	newwindow.focus();
	return newwindow;
}


function doNamedPopup(name,url,width,height)
{
	newwindow = window.open(url,name,"width="+width+",height="+height+",toolbar=0,location=0,directories=0,status=1,menubar=1,scrollbars=1,resizable=1");
	newwindow.focus();
	return newwindow;
}

function setCookie( name, value, days )
{
	var exp = new Date();
	var days2Live = exp.getTime() + (24 * 60 * 60 * 1000 * days);
	exp.setTime(days2Live);
	if( document.location.href.indexOf(".real.com") >=0 )
		var domname=".real.com";
	else
		var domname=".prognet.com";
	document.cookie = name+"="+value+"; expires=" + exp.toGMTString() + "; domain=" + domname+"; path=/";
}

function setSessionCookie( name, value )
{
	if( document.location.href.indexOf(".real.com") >=0 )
		var domname=".real.com";
	else
		var domname=".prognet.com";
	document.cookie = name+"="+value+"; domain=" + domname+"; path=/";
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
	var pagetype = ( dualProx.arguments.length > 3 )?dualProx.arguments[3]:"non";
	var width = ( dualProx.arguments.length > 4 )?dualProx.arguments[4]:"200";
	var height = ( dualProx.arguments.length > 5 )?dualProx.arguments[5]:"200";
	var clienttype = ( dualProx.arguments.length > 6 )?dualProx.arguments[6]:"";
	if( type == "embedded" )
		var newWin=window.open(mediaurl,newWin);
	else
		document.location.href= mediaurl;
	var modeStr = "";
	var x;
	if( clienttype == "useplayer" )
	{
		modeStr = "action=_tunerplayer";
	}else{
		modeStr = "mode=compact";
	}
	
	if( clienttype == "useplayer" || clienttype == "usejukebox" )
	{
		if( document.location.href.indexOf("?") >= 0 )
		{
			document.location.href = mediaurl + "&" + modeStr;
		}else{
			document.location.href = mediaurl + "?" + modeStr;
		}
		
	}	
	timerID = setTimeout("secondSpawn('"+pageurl+"','"+pagetype+"','"+width+"','"+height+"')",5000);
	timerRunning = true;
}

function secondSpawn(url) 
{
	var pagetype = ( secondSpawn.arguments.length > 1 )?secondSpawn.arguments[1]:"non";
		
	if( pagetype == "popup" )
		var popWin=window.open(url,popWin);
	else
		window.self.location=url;
}
