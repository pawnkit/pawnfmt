stock FloatLiterals()
{
    new Float:tiny = 1.0E-3;
    new Float:huge = 2.5E+10;
    new Float:avogadro = 6.02E23;
    new Float:half = 0.5;
    return Float:(tiny + huge + avogadro + half);
}
