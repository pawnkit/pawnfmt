stock SizeofPatterns(values[][], playerid)
{
    new total = sizeof values;
    new row = sizeof values[playerid];
    new wrapped = sizeof(values[playerid]);
    format(values[playerid], sizeof values[playerid], "%c", 'A');
    return total + row + wrapped;
}
