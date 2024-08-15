// MIT License

// Copyright (c) 2024 Saeed Rasooli
// Copyright (c) 2012-2020 Mat Ryer, Tyler Bunnell and contributors.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package require

import (
	"cmp"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ilius/is/v2"
)

type PanicTestFunc func()
type Comparison func() (success bool)

type TestingT = testing.TB

func addMsg(is *is.Is, msgAndArgs []any) {
	if len(msgAndArgs) > 0 {
		format := msgAndArgs[0].(string)
		is.AddMsg(format, msgAndArgs[1:]...)
	}
}

func Condition(t TestingT, comp Comparison, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.True(comp())
}

func Conditionf(t TestingT, comp Comparison, msg string, args ...any) {
	is := is.New(t)
	is.AddMsg(msg, args...)
	is.True(comp())
}

func Contains(t TestingT, s any, contains any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Contains(s, contains)
}

func Containsf(t TestingT, s any, contains any, msg string, args ...any) {
	is := is.New(t)
	is.AddMsg(msg, args...)
	is.Contains(s, contains)
}

func ElementsMatch(t TestingT, listA any, listB any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	if isEmpty(listA) && isEmpty(listB) {
		return
	}
	if !isList(listA) {
		is.Fail(fmt.Sprintf("%q has an unsupported type %T, expecting array or slice", listA, listA))
		return
	}
	if !isList(listB) {
		is.Fail(fmt.Sprintf("%q has an unsupported type %T, expecting array or slice", listB, listB))
		return
	}
	extraA, extraB := diffLists(listA, listB)

	if len(extraA) == 0 && len(extraB) == 0 {
		return
	}
	is.Fail(fmt.Sprintf("lists are not equal, %d extra in first, %d extra in second", len(extraA), len(extraB)))
}

func ElementsMatchf(t TestingT, listA any, listB any, msg string, args ...any) {
	ElementsMatch(t, listA, listB, append([]any{msg}, args...)...)
}

func Empty(t TestingT, object any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	if !isEmpty(object) {
		is.Fail(fmt.Sprintf("Should be empty, but was %v", object))
	}
}

func Equal(t TestingT, expected any, actual any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Equal(actual, expected)
}

func EqualError(t TestingT, theError error, errString string, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.ErrMsg(theError, errString)
}

func EqualErrorf(t TestingT, theError error, errString string, msg string, args ...any) {
	EqualError(t, theError, errString, append([]any{msg}, args...)...)
}

func EqualExportedValues(t TestingT, expected any, actual any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)

	aType := reflect.TypeOf(expected)
	bType := reflect.TypeOf(actual)

	if aType != bType {
		is.Fail(fmt.Sprintf("Types expected to match exactly\n\t%v != %v", aType, bType))
		return
	}

	if aType.Kind() == reflect.Ptr {
		aType = aType.Elem()
	}
	if bType.Kind() == reflect.Ptr {
		bType = bType.Elem()
	}

	if aType.Kind() != reflect.Struct {
		is.Fail(fmt.Sprintf("Types expected to both be struct or pointer to struct \n\t%v != %v", aType.Kind(), reflect.Struct))
		return
	}

	if bType.Kind() != reflect.Struct {
		is.Fail(fmt.Sprintf("Types expected to both be struct or pointer to struct \n\t%v != %v", bType.Kind(), reflect.Struct))
		return
	}

	expected = copyExportedFields(expected)
	actual = copyExportedFields(actual)

	if !objectsAreEqualValues(expected, actual) {
		// diff := diff(expected, actual)
		// expected, actual = formatUnequalValues(expected, actual)
		is.Fail(fmt.Sprintf(
			"Not equal (comparing only exported fields): \nexpected: %s\nactual  : %s",
			expected, actual,
			// diff,
		))
	}
}

func EqualExportedValuesf(t TestingT, expected any, actual any, msg string, args ...any) {
	EqualExportedValues(t, expected, actual, append([]any{msg}, args...)...)
}

func EqualValues(t TestingT, expected any, actual any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Equal(actual, expected)
	is.EqualType(expected, actual)
}

func EqualValuesf(t TestingT, expected any, actual any, msg string, args ...any) {
	EqualValues(t, expected, actual, append([]any{msg}, args...)...)
}

func Equalf(t TestingT, expected any, actual any, msg string, args ...any) {
	Equal(t, expected, actual, append([]any{msg}, args...)...)
}

func Error(t TestingT, err error, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Err(err)
}

// ErrorAs asserts that at least one of the errors in err's chain matches target, and if so, sets target to that error value.
// This is a wrapper for errors.As.
func ErrorAs(t TestingT, err error, target any, msgAndArgs ...any) {
	// TODO
	is := is.New(t)
	is.Fail("unsupported function")
}

func ErrorAsf(t TestingT, err error, target any, msg string, args ...any) {
	ErrorAs(t, err, target, append([]any{msg}, args...)...)
}

func ErrorContains(t TestingT, theError error, contains string, msgAndArgs ...any) {
	// TODO
	is := is.New(t)
	is.Fail("unsupported function")
}

func ErrorContainsf(t TestingT, theError error, contains string, msg string, args ...any) {
	ErrorContains(t, theError, contains, append([]any{msg}, args...)...)
}

// ErrorIs asserts that at least one of the errors in err's chain matches target.
// This is a wrapper for errors.Is.
func ErrorIs(t TestingT, err error, target error, msgAndArgs ...any) {
	// TODO
	is := is.New(t)
	is.Fail("unsupported function")
}

func ErrorIsf(t TestingT, err error, target error, msg string, args ...any) {
	ErrorIs(t, err, target, append([]any{msg}, args...)...)
}

func Errorf(t TestingT, err error, msg string, args ...any) {
	Error(t, err, append([]any{msg}, args...)...)
}

func Eventually(t TestingT, condition func() bool, waitFor time.Duration, tick time.Duration, msgAndArgs ...any) {
	// TODO
	is := is.New(t)
	is.Fail("unsupported function")
}

func EventuallyWithT(t TestingT, condition func(collect TestingT), waitFor time.Duration, tick time.Duration, msgAndArgs ...any) {
	// TODO
	is := is.New(t)
	is.Fail("unsupported function")
}

func EventuallyWithTf(t TestingT, condition func(collect TestingT), waitFor time.Duration, tick time.Duration, msg string, args ...any) {
	EventuallyWithT(t, condition, waitFor, tick, append([]any{msg}, args...)...)
}

func Eventuallyf(t TestingT, condition func() bool, waitFor time.Duration, tick time.Duration, msg string, args ...any) {
	Eventually(t, condition, waitFor, tick, append([]any{msg}, args...)...)
}

func Exactly(t TestingT, expected any, actual any, msgAndArgs ...any) {
	is := is.New(t)
	is.Equal(actual, expected)
}

func Exactlyf(t TestingT, expected any, actual any, msg string, args ...any) {
	Exactly(t, expected, actual, append([]any{msg}, args...)...)
}

func Fail(t TestingT, failureMessage string, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Fail(failureMessage)
}

func FailNow(t TestingT, failureMessage string, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Fail(failureMessage)
}

func FailNowf(t TestingT, failureMessage string, msg string, args ...any) {
	FailNow(t, failureMessage, append([]any{msg}, args...)...)
}

func Failf(t TestingT, failureMessage string, msg string, args ...any) {
	is := is.New(t)
	is.AddMsg(msg, args...)
	is.Fail(failureMessage)
}

func False(t TestingT, value bool, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.False(value)
}

func Falsef(t TestingT, value bool, msg string, args ...any) {
	False(t, value, append([]any{msg}, args...)...)
}

func FileExists(t TestingT, path string, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			is.Fail(fmt.Sprintf("unable to find file %q", path))
			return
		}
		is.Fail(fmt.Sprintf("error when running os.Lstat(%q): %s", path, err))
		return
	}
	if info.IsDir() {
		is.Fail(fmt.Sprintf("%q is a directory", path))
		return
	}
}

func Greater[T cmp.Ordered](t TestingT, e1 T, e2 T, msgAndArgs ...any) {
	if e1 > e2 {
		return
	}
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Fail(fmt.Sprintf("\"%v\" is not greater than \"%v\"", e1, e2))
}

func GreaterOrEqual[T cmp.Ordered](t TestingT, e1 T, e2 T, msgAndArgs ...any) {
	if e1 >= e2 {
		return
	}
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Fail(fmt.Sprintf("\"%v\" is not greater than or equal to \"%v\"", e1, e2))
}

func GreaterOrEqualf[T cmp.Ordered](t TestingT, e1 T, e2 T, msg string, args ...any) {
	GreaterOrEqual(t, e1, e2, append([]any{msg}, args...)...)
}

func Greaterf[T cmp.Ordered](t TestingT, e1 T, e2 T, msg string, args ...any) {
	Greater(t, e1, e2, append([]any{msg}, args...)...)
}

func HTTPBodyContains(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, str any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func HTTPBodyContainsf(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, str any, msg string, args ...any) {
	HTTPBodyContains(t, handler, method, url, values, str, append([]any{msg}, args...)...)
}

func HTTPBodyNotContains(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, str any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func HTTPBodyNotContainsf(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, str any, msg string, args ...any) {
	HTTPBodyNotContains(t, handler, method, url, values, str, append([]any{msg}, args...)...)
}

func HTTPError(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func HTTPErrorf(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, msg string, args ...any) {
	HTTPError(t, handler, method, url, values, append([]any{msg}, args...)...)
}

func HTTPRedirect(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func HTTPRedirectf(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, msg string, args ...any) {
	HTTPRedirect(t, handler, method, url, values, append([]any{msg}, args...)...)
}

func HTTPStatusCode(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, statuscode int, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func HTTPStatusCodef(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, statuscode int, msg string, args ...any) {
	HTTPStatusCode(t, handler, method, url, values, statuscode, append([]any{msg}, args...)...)
}

func HTTPSuccess(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func HTTPSuccessf(t TestingT, handler http.HandlerFunc, method string, url string, values url.Values, msg string, args ...any) {
	HTTPSuccess(t, handler, method, url, values, append([]any{msg}, args...)...)
}

func Implements(t TestingT, interfaceObject any, object any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
}

func Implementsf(t TestingT, interfaceObject any, object any, msg string, args ...any) {
	Implements(t, interfaceObject, object, append([]any{msg}, args...)...)
}

func NoFileExists(t TestingT, path string, msgAndArgs ...interface{}) bool {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	info, err := os.Lstat(path)
	if err != nil {
		return true
	}
	if info.IsDir() {
		return true
	}
	is.Fail(fmt.Sprintf("file %q exists", path))
	return false
}

func DirExists(t TestingT, path string, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			is.Fail(fmt.Sprintf("unable to find file %q", path))
			return
		}
		is.Fail(fmt.Sprintf("error when running os.Lstat(%q): %s", path, err))
		return
	}
	if !info.IsDir() {
		is.Fail(fmt.Sprintf("%q is a file", path))
	}
}

// NoDirExists checks whether a directory does not exist in the given path.
// It fails if the path points to an existing _directory_ only.
func NoDirExists(t TestingT, path string, msgAndArgs ...interface{}) bool {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		return true
	}
	if !info.IsDir() {
		return true
	}
	is.Fail(fmt.Sprintf("directory %q exists", path))
	return false
}

func DirExistsf(t TestingT, path string, msg string, args ...any) {
	DirExists(t, path, append([]any{msg}, args...)...)
}

func JSONEq(t TestingT, expected string, actual string, msgAndArgs ...interface{}) bool {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
	return false
}

func YAMLEq(t TestingT, expected string, actual string, msgAndArgs ...interface{}) bool {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	// TODO
	is.Fail("unsupported function")
	return false
}

func IsType(t TestingT, expectedType any, object any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.IsType(expectedType.(reflect.Type), object)
}

func Len(t TestingT, object any, length int, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Len(object, length)
}

func Nil(t TestingT, object any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.Nil(object)
}

func NoError(t TestingT, err error, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.NotErr(err)
}

func NotNil(t TestingT, object any, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.NotNil(object)
}

func Panics(t TestingT, f PanicTestFunc, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.ShouldPanic(f)
}

func True(t TestingT, value bool, msgAndArgs ...any) {
	is := is.New(t)
	addMsg(is, msgAndArgs)
	is.True(value)
}
