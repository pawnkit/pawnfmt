stock Float:CastPatterns(value,offset,values[])
{
	new Float:scaled=Float:(values[offset]+value);
	new cell=_:scaled;
	new Float:sum=Float:(cell+value)*Float:(values[offset]);
	return Float:(sum+Float:(cell));
}