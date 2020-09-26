#include "CppUTest/CommandLineTestRunner.h"
#include "CppUTest/TestHarness.h"
#include "CppUTestExt/MockSupport.h"

#include <iostream>

#include "foo.h"
#include "mock_foo.cpp"

TEST_GROUP(TestMain)
{
    TEST_SETUP()
    {
    }

    TEST_TEARDOWN()
    {
        mock().clear();
    }
};

TEST(TestMain, TestFooSuccess)
{
    expect_foo(0, 1, 2);
    int result = foo(0, 1);
    CHECK_EQUAL(2, result);

    mock().checkExpectations();
}

TEST(TestMain, TestPiyoSuccess)
{
    expect_piyo(0, 1, 2);
    int result = piyo(0, 1);
    CHECK_EQUAL(2, result);

    mock().checkExpectations();
}

int main(int argc, char **argv)
{
    return CommandLineTestRunner::RunAllTests(argc, argv);
}
