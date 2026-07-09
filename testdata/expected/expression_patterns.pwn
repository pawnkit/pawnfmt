stock EvaluateExpressions(value, a, b, c, offset)
{
    new flags = value & ~(1 << offset);
    new packed = (a << 4) | (b >> 2);
    new mixed = a ^ b & c;
    new status = a ? GetValue(a) + offset : c;
    new nested = a > 0 ? (b < c ? a : b) : c;
    return flags + packed + mixed + status + nested;
}

stock Float:Clamp(Float:value)
{
    return Float:(value < Float:(0.0) ? Float:(0.0) : value);
}

stock Float:Invoke()
{
    return Float:foo();
}

stock _:AsCell(Float:x)
{
    return _:x;
}
