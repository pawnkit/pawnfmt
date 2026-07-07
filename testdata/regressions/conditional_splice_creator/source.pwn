stock F(playerid, keys, PX, PY, PZ, cX, cY, cZ, rX, rY, rZ, obj_x, obj_y, obj_z, minZ, tmp) {
#if defined CA_RayCastLineAngle\
				&& defined CA_GetModelBoundingBox
				if(CA_RayCastLineAngle(PX, PY, PZ, obj_x, obj_y, obj_z, cX, cY, cZ, rX, rY, rZ))
				{
					CA_GetModelBoundingBox(ObjectsInfo[CreatorInfo[playerid][ucHoldingObj]][oModelid], tmp, tmp, minZ, tmp, tmp, tmp);
					#if defined SetDynamicObjectPos\
						&& defined GetDynamicObjectRot\
						&& defined SetDynamicObjectRot
						SetDynamicObjectPos(CreatorInfo[playerid][ucHoldingObj], cX, cY, cZ + floatabs(minZ));
						if(keys & KEY_WALK && keys & KEY_JUMP) GetDynamicObjectRot(CreatorInfo[playerid][ucHoldingObj], obj_x, obj_y, rZ);
						else GetDynamicObjectRot(CreatorInfo[playerid][ucHoldingObj], rX, rY, rZ);
						SetDynamicObjectRot(CreatorInfo[playerid][ucHoldingObj], rX, rY, rZ);
					#else
						SetObjectPos(CreatorInfo[playerid][ucHoldingObj], cX, cY, cZ + floatabs(minZ));
						if(keys & KEY_WALK && keys & KEY_JUMP) GetObjectRot(CreatorInfo[playerid][ucHoldingObj], obj_x, obj_y, rZ);
						else GetObjectRot(CreatorInfo[playerid][ucHoldingObj], rX, rY, rZ);
						SetObjectRot(CreatorInfo[playerid][ucHoldingObj], rX, rY, rZ);
					#endif
				}
				else
				{
			#endif
				#if defined SetDynamicObjectPos
					SetDynamicObjectPos(CreatorInfo[playerid][ucHoldingObj], obj_x, obj_y, obj_z);
				#else
					SetObjectPos(CreatorInfo[playerid][ucHoldingObj], obj_x, obj_y, obj_z);
				#endif
			#if defined CA_RayCastLineAngle\
				&& defined CA_GetModelBoundingBox
				}
			#endif
}
