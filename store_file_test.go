
// const (
// 	x200 = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
// 	x201 = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
// )
// 
// func TestUrlSafename(tt *testing.T) {
// 	t := (*T)(tt)
// 
// 	// Test that different URIs end up generating different safe names
// 	t.assertEQ("example.org,fred,a=b,58489f63a7a83c3b7794a6a398ee8b1f", safeName("http://example.org/fred/?a=b"))
// 	t.assertEQ("example.org,fred,a=b,8c5946d56fec453071f43329ff0be46b", safeName("http://example.org/fred?/a=b"))
// 	t.assertEQ("www.example.org,fred,a=b,499c44b8d844a011b67ea2c015116968", safeName("http://www.example.org/fred?/a=b"))
// 	t.assertEQ("www.example.org,fred,a=b,692e843a333484ce0095b070497ab45d", safeName("https://www.example.org/fred?/a=b"))
// 	t.assertNE(safeName("http://www"), safeName("https://www"))
// 
// 	// Test the max length limits
// 	uri := "http://" + x200 + ".org"
// 	uri2 := "http://" + x201 + ".org"
// 	t.assertNE(safeName(uri2), safeName(uri))
// 	// Max length should be 200 + 1 (",") + 32
// 	t.want(233 == len(safeName(uri)), "max len 233, but got %d", len(safeName(uri)))
// 	t.want(233 == len(safeName(uri2)), "max len 233, but got %d", len(safeName(uri2)))
// 
// 	// Unicode
// 	t.assertEQ("xn--http,-4y1d.org,fred,a=b,579924c35db315e5a32e3d9963388193", safeName("http://\u2304.org/fred/?a=b"))
// }

