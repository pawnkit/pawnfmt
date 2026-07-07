new Float:matrix[2][2]={{1.0,2.0},{3.0,4.0}},bool:states[2]={true,false};

stock UseMixedDecls()
{
	return matrix[0][0]+states[0];
}