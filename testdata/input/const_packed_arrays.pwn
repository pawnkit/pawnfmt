const names[2][8 char]={!"A",!"BC"};
static const bool:flags[2]={true,false};

stock ConstPacked()
{
	return names[0][0]+flags[0];
}