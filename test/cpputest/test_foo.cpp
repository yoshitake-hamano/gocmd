#include "CppUTest/CommandLineTestRunner.h"
#include <iostream>
#include "foo.h"

TEST_GROUP(TestFoo)
{
    TEST_SETUP()
    {
    }

    TEST_TEARDOWN()
    {
    }
};

TEST(TestFoo, TestSuccess)
{
}

int main(int argc, char **argv)
{
    return CommandLineTestRunner::RunAllTests(argc, argv);
}
