stock GroupedMath(value,data[])
{
	new grouped=((value+1)*(data[0]-2));
	new mixed=(value+(data[1]<<2))^(data[2]&7);
	return grouped+mixed;
}