forward OnDialogMessage(playerid,dialogid,const format[],{Float,_}:...)<default,connected,idle>;

public OnDialogMessage(playerid,dialogid,const format[],{Float,_}:...)<connected,idle>
{
	return dialogid+playerid;
}