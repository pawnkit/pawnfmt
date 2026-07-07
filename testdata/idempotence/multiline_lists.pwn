forward VeryLongCallbackName(playerid, const playerName[], Float:spawnX, Float:spawnY, Float:spawnZ);

stock CallLongThing(playerid)
{
    SendClientMessage(playerid, 0xFFFFFFFF, "hello from a callback-heavy game mode");
    return 1;
}