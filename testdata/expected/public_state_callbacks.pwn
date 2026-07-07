forward OnStateEvent(playerid) <default, connected>;

public OnStateEvent(playerid) <default>
{
    return 1;
}

public OnStateEvent(playerid) <connected, idle>
{
    return playerid;
}
