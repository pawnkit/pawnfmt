new gPlayerCount;
new Float:gSpawnX=123.5;
new bool:gConnected[MAX_PLAYERS];
const Float:gScale=Float:(1.0);
static const gFlags=1;
public numHaleKills=0;
stock Float:gSpawnY=1.0;
public Float:gPos[3]={1.0,2.0,3.0};

forward Foo(...);
forward SetSpawn(Float:x,Float:y,Float:z=0.0);
forward SendClientMessagef(playerid,color,const format[],{Float,_}:...);
native GetPlayerPos(playerid,&Float:x,&Float:y,&Float:z);
native Bar(const name[],&Float:x=1.0);
forward Float:GetSpawnX(playerid);
native SetPlayerPos(playerid,Float:x,Float:y,Float:z);
native Float:GetVectorLength(Float:x,Float:y)=VectorSize;

new Float:x=1.0,y[4],bool:z=true;
new _:value=1;