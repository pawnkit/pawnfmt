public OnPlayerCommandText(playerid, cmdtext[])
{
    new message[144 char];
    format(message, sizeof message, "player %d typed %s", playerid, cmdtext);
    SendClientMessage(playerid, 0xFFFFFFFF, message);
    return 1;
}