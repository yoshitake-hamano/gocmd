#include <stdio.h>

const int globalConstInt = 0x414243;
int globalInt = 0xAB;
char *globalChar = "global variable";

static void example()
{
    char *local = "local variable";
    printf("Hello World\n");
    printf("HELLO WORLD\n");
    printf("local  : %s\n", local);
    printf("global : %s\n", globalChar);
    printf("global const int : %d\n", globalConstInt);
    globalInt = 0xCD;
    printf("global const int : %d\n", globalInt);
    printf("%s %s:%d\n", __FILE__, __FUNCTION__, __LINE__);
}

int main()
{
    example();
    return 0;
}
