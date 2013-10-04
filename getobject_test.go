/**
	provides the photo extraction core testing
 */
package gorets

import (
	"bytes"
	"net/http"
	"io/ioutil"
	"testing"
	"net/textproto"
)

func TestGetObject(t *testing.T) {
	header := http.Header{}
	textproto.MIMEHeader(header).Add("Content-Type","image/jpeg")
	textproto.MIMEHeader(header).Add("Content-ID", "123456")
	textproto.MIMEHeader(header).Add("Object-ID","1")
	textproto.MIMEHeader(header).Add("Preferred", "1")
	textproto.MIMEHeader(header).Add("UID", "1a234234234")
	textproto.MIMEHeader(header).Add("Description", "Outhouse")
	textproto.MIMEHeader(header).Add("Sub-Description", "The urinal")
	textproto.MIMEHeader(header).Add("Location", "http://www.simpleboundary.com/image-5.jpg")

	var body string = `<binary data 1>`
	reader := ioutil.NopCloser(bytes.NewReader([]byte(body)))

	objects, err := parseGetObjectResult(header, reader)
	if err != nil {
		t.Error("error parsing multipart: "+ err.Error())
	}

	counter := 0
	o := <- objects.Objects
	if !o.Preferred {
		t.Errorf("error parsing preferred at object %d", counter)
	}
	AssertEquals(t, "bad value", "image/jpeg", o.ContentType)
	AssertEquals(t, "bad value", "123456", o.ContentId)
	AssertEqualsInt(t, "bad value", 1, o.ObjectId)
	AssertEquals(t, "bad uid", "1a234234234", o.Uid)
	AssertEquals(t, "bad value", "Outhouse", o.Description)
	AssertEquals(t, "bad value", "The urinal", o.SubDescription)
	AssertEquals(t, "bad value", "<binary data 1>", string(o.Blob))
	AssertEquals(t, "bad value", "http://www.simpleboundary.com/image-5.jpg", o.Location)
	if o.RetsError {
		t.Errorf("error parsing rets error at object %d", counter)
	}
}

var boundary string = "simple boundary"

var contentType string = `multipart/parallel; boundary="simple boundary"`

var multipartBody string =
	`--simple boundary
Content-Type: image/jpeg
Content-ID: 123456
Object-ID: 1
Preferred: 1

<binary data 1>
--simple boundary
Content-Type: image/jpeg
Content-ID: 123456
Object-ID: 2
UID: 1a234234234

<binary data 2>
--simple boundary
Content-Type: image/jpeg
Content-ID: 123456
Object-ID: 3
Description: Outhouse
Sub-Description: The urinal

<binary data 3>
--simple boundary
Content-Type: text/xml
Content-ID: 123457
Object-ID: 4
RETS-Error: 1

<RETS ReplyCode="20403" ReplyText="There is no object with that
Object-ID"/>

--simple boundary
Content-Type: image/jpeg
Content-ID: 123456
Object-ID: 5
Location: http://www.simpleboundary.com/image-5.jpg

<binary data 5>
--simple boundary--`

func TestExtractBoundary(t *testing.T) {
	extracted := extractBoundary(contentType)

	AssertEquals(t, "bad boundary", boundary, extracted)
}

func TestGetObjects(t *testing.T) {
	extracted := extractBoundary(contentType)

	AssertEquals(t, "bad boundary", boundary, extracted)

	body := ioutil.NopCloser(bytes.NewReader([]byte(multipartBody)))

	objects, err := parseGetObjectsResult(extracted, body)
	if err != nil {
		t.Error("error parsing multipart: "+ err.Error())
	}

	counter := 0
	o1 := <- objects.Objects
	if !o1.Preferred {
		t.Errorf("error parsing preferred at object %d", counter)
	}
	AssertEquals(t, "bad value", "image/jpeg", o1.ContentType)
	AssertEquals(t, "bad value", "123456", o1.ContentId)
	AssertEqualsInt(t, "bad value", 1, o1.ObjectId)
	AssertEquals(t, "bad value", "<binary data 1>", string(o1.Blob))

	o2 := <- objects.Objects
	AssertEqualsInt(t, "bad value", 2, o2.ObjectId)
	AssertEquals(t, "bad uid", "1a234234234", o2.Uid)

	o3 := <- objects.Objects
	AssertEqualsInt(t, "bad value", 3, o3.ObjectId)
	AssertEquals(t, "bad value", "Outhouse", o3.Description)
	AssertEquals(t, "bad value", "The urinal", o3.SubDescription)

	o4 := <- objects.Objects
	if !o4.RetsError {
		t.Errorf("error parsing error at object %d", counter)
	}
	AssertEquals(t, "bad value", "text/xml", o4.ContentType)

	if objects.ProcessingFailure != nil {
		t.Errorf("error parsing body at object %d: %s", counter, objects.ProcessingFailure.Error())
	}

	o5 := <- objects.Objects
	AssertEquals(t, "bad value", "http://www.simpleboundary.com/image-5.jpg", o5.Location)
	AssertEquals(t, "bad value", "image/jpeg", o5.ContentType)
	AssertEquals(t, "bad value", "123456", o5.ContentId)
	AssertEqualsInt(t, "bad value", 5, o5.ObjectId)
	AssertEquals(t, "bad value", "<binary data 5>", string(o5.Blob))

	if objects.ProcessingFailure != nil {
		t.Errorf("error parsing body at object %d: %s", counter, objects.ProcessingFailure.Error())
	}

}
