stock InitFromCalls(playerid, names[][])
{
    new length = strlen(names[playerid]);
    new first = names[playerid][0];
    new bool:ready = IsPlayerConnected(playerid);
    return length + first + ready;
}
