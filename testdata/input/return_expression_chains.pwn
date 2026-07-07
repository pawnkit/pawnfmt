stock ReturnChains(values[][],index,playerid)
{
	return values[GetSlot(playerid)][index[0]++] + Adjust(values[index[1]][0],index[1]--);
}