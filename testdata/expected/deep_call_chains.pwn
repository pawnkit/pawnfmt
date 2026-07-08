stock DeepCallChains(values[][], idx, playerid)
{
    values[idx][0] = Compose(
        GetMatrix()[idx][FindSlot(playerid)],
        Resolve(values[NextIndex(idx)][0]),
        Build(values[idx][1], playerid)
    );
    return Dispatch(
        GetMatrix()[idx][0],
        Resolve(values[idx][1]),
        Build(values[idx][2], FindSlot(playerid))
    );
}
