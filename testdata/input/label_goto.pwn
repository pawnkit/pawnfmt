stock TestLabel(value)
{
    if (value < 0)
        goto invalid_value;

    return 1;

invalid_value:
  return 0;

next_step: goto invalid_value;
}