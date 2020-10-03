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
    class Test{
    public:
        int lhd;
        int rhd;
        int retval;

    public:
        Test(int lhd, int rhd, int retval) : lhd(lhd), rhd(rhd), retval(retval) {}
    };
    Test tests[] = {
        Test(0, 1, 2),
        Test(2, 1, 3),
        Test(2, 2, 3),
    };
    int length = sizeof(tests)/sizeof(tests[0]);
    for (int i=0; i<length; i++) {
        Test *t = &tests[i];
        expect_piyo(t->lhd, t->rhd, t->retval);
        int result = piyo(t->lhd, t->rhd);
        CHECK_EQUAL(t->retval, result);

        mock().checkExpectations();
        mock().clear();
    }
}

int main(int argc, char **argv)
{
    return CommandLineTestRunner::RunAllTests(argc, argv);
}
