stock LoopHeaders(limit,items[])
{
	for(new i=0,j=limit;i<j;i++,j--)
		items[i]+=j;

	for(i=0,j=limit-1;i<j;i++,j--){
		items[i]=items[j];
	}

	for(;limit>0;limit--)
		items[0]+=limit;

	return items[0];
}