stock OperatorPrecedence(a, b, c, d, flags, mask)
{
    new mixed = a << 2 | b >> 1 & c ^ d;
    new filtered = (flags & ~mask) | 1 << c;
    new grouped = (a | b) << (c & d);
    new compared = a < b & c ^ d;
    return mixed + filtered + grouped + compared;
}
