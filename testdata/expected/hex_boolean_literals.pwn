stock LiteralKinds(flags)
{
    new mask = 0xFF00AA11;
    new low = 0x10;
    new bool: enabled = false;
    if (enabled)
    {
        return mask;
    }
    return (mask & low) | flags;
}
