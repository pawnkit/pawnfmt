stock MoreInlineLabels(value)
{
    setup: if (value > 0) {
        return 1;
    }

    fallback: new next = 1;
    loop_head: while (next) break;
    branch: switch (value) { default: return 0; }
}