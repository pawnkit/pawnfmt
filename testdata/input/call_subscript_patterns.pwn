stock CallSubscriptPatterns(playerid,cmdtext[],labels[][])
{
	if(!strcmp(cmdtext,"/spawn",true,sizeof cmdtext))
		return 1;

	format(labels[playerid],sizeof labels[playerid],"%s",labels[playerid]);

	new first=labels[playerid][0];
	labels[playerid][0]='X';

	return sizeof labels[playerid] + first;
}