stock ArgSplit(playerid)
{
    SendClientMessage(playerid,
#if defined DEBUG
        0xFFFFFFFF,
#else
        0x00FF00FF,
#endif
        "hello");
}