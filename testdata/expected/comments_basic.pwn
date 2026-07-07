// Describes the basic comment-handling test cases.
#include <a_samp>

// Called when a player connects.
public OnPlayerConnect(playerid)
{
    // greet the player
    new msg[32] = "hi"; // reuse this buffer below
    SendClientMessage(playerid, -1, msg);
    /* nothing else to do here */
    return 1;
}

stock Plain(x, y)
{
    new z = x + y;
    return z;
}
