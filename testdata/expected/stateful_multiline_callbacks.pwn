forward OnDialogFlow(
    playerid,
    dialogid,
    response,
    listitem,
    inputtext[]
) <default, connected, idle>;

public OnDialogFlow(playerid, dialogid, response, listitem, inputtext[]) <connected, idle>
{
    return response + listitem + playerid;
}
