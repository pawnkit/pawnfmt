new bool:gFlags[MAX_PLAYERS] = {true, ...};
new matrix[2][2] = {{1, 2}, {3, 4}};
new packedTable[3][16 char] = {!"A", !"BC", ...};
new Float:packedTagged[2][8 char] = {!"x", !"yz"};

stock PackedEdge()
{
    label_a:
    new names[3][8 char] = {!"A", ..., !"XYZ"};
    goto label_a;
}
