stock ComputedDimensions(values[][],playerid)
{
	new row[sizeof values[playerid]];
	new masks[(1<<3)|1];
	new labels[2][(1<<4) char];
	return sizeof row+sizeof masks+sizeof labels[playerid];
}