static labels[2][16 char]={!"Alpha",!"Beta"};
static bool:readyStates[2];

stock StaticPacked()
{
	return labels[0][0]+readyStates[0];
}