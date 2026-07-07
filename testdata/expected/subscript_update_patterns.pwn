stock SubscriptOps(values[][], index, next)
{
    new current = values[index[0]][index[1]];
    values[index[0]][index[1]] = values[next[0]][next[1]];
    values[index[0]][next[1]]++;
    --values[next[0]][index[1]];
    return values[GetSlot()][index[0]++] + values[next[0]][--next[1]];
}
