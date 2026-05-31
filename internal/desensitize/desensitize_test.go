package desensitize

import "testing"

func TestBuiltInRules(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"user", Desensitized("100", UserID), "0"},
		{"name", ChineseName("段正淳"), "段**"},
		{"id", IDCardNum("51343620000320711X", 1, 2), "5***************1X"},
		{"fixed", FixedPhone("09157518479"), "0915*****79"},
		{"mobile", MobilePhone("18049531999"), "180****1999"},
		{"address", Address("北京市海淀区马连洼街道289号", 8), "北京市海淀区马********"},
		{"email", Email("duandazhi-jack@gmail.com.cn"), "d*************@gmail.com.cn"},
		{"password", Password("1234567890"), "**********"},
		{"car7", CarLicense("苏D40000"), "苏D4***0"},
		{"car8", CarLicense("陕A12345D"), "陕A1****D"},
		{"bank", BankCard("11011111222233333256"), "1101 **** **** **** 3256"},
		{"ipv4", IPv4("192.168.1.1"), "192.*.*.*"},
		{"ipv6", IPv6("2001:0db8:86a3:08d3:1319:8a2e:0370:7344"), "2001:*:*:*:*:*:*:*"},
		{"passport", Passport("PJ1234567"), "PJ*****67"},
		{"credit", CreditCode("91110108MA01ABCDE7"), "9111**********CDE7"},
		{"first", FirstMask("123456789"), "1********"},
		{"clear", Clear(), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q want %q", tc.got, tc.want)
			}
		})
	}
}

func TestDesensitizedPtrAndBoundary(t *testing.T) {
	if got := Desensitized("18049531999", MobilePhoneType); got != "180****1999" {
		t.Fatalf("Desensitized: %q", got)
	}
	if DesensitizedPtr("x", ClearToNullType) != nil {
		t.Fatal("ClearToNullType should return nil pointer")
	}
	if got := IDCardNum("123", 2, 2); got != "" {
		t.Fatalf("invalid id mask: %q", got)
	}
	if got := BankCard("1234 5678"); got != "12345678" {
		t.Fatalf("short bank card: %q", got)
	}
}
