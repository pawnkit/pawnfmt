stock NestedArgs(values[][], index, playerid, buffer[])
{
    EmitValue(playerid, values[index[0]][index[1]], GetSlot(index), buffer[FindOffset(playerid)]);
    return Compose(values[index[0]][0], buffer, index[0]++);
}
