stock UnaryPatterns(flags,value,data[])
{
	new masked=~(flags<<1);
	new negated=-(value+data[0]);
	new shifted=~data[value-1];
	return masked+negated+shifted;
}