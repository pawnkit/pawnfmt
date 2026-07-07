stock NestedControlFlow(limit, flags, values[])
{
    for (new i = 0; i < limit; i++)
    {
        while (values[i] > 0)
        {
            switch (values[i])
            {
                case 1:
                    {
                        values[i] -= 1;
                        break;
                    }
                case 2:
                    if (flags)
                    {
                        continue;
                    }
                default:
                    values[i]--;
            }
            if (values[i] == 3)
            {
                break;
            }
            values[i]--;
        }
    }
    return values[0];
}
