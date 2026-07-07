stock StringLiterals(playerid)
{
    new welcome[] = "Hello\nWorld";
    new path[] = "C:\\pawn\\script";
    SendClientMessage(playerid, 0xFFFFFFFF, "tab:\tpath");
    return welcome[0] + path[0];
}
