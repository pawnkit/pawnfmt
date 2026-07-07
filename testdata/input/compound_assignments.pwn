stock CompoundOps(value, flags, mask, shift, data[])
{
	value+=10;
	value-=5;
	value*=2;
	value/=4;
	value%=3;
	flags&=~mask;
	flags|=1<<shift;
	flags^=data[0];
	data[shift]+=100;
	data[shift>>1]-=flags;
	data[0]<<=1;
	data[1]>>=2;
	return value+flags+data[shift];
}