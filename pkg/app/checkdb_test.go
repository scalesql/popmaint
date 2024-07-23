package app

// func TestDatabaseSort(t *testing.T) {
// 	assert := assert.New(t)
// 	dd := []mssqlz.Database{
// 		{
// 			DatabaseName: "One",
// 			LastDBCC:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			DatabaseMB:   100,
// 		},
// 		{
// 			DatabaseName: "Two",
// 			LastDBCC:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
// 			DatabaseMB:   300,
// 		},
// 		{
// 			DatabaseName: "Three",
// 			LastDBCC:     time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
// 			DatabaseMB:   100,
// 		},
// 		{
// 			DatabaseName: "Four",
// 			LastDBCC:     time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC),
// 			DatabaseMB:   200,
// 		},
// 	}
// 	sortDatabases(dd)
// 	assert.Equal("Four", dd[0].DatabaseName)
// 	assert.Equal("Three", dd[1].DatabaseName)
// 	assert.Equal("Two", dd[2].DatabaseName)
// 	assert.Equal("One", dd[3].DatabaseName)
// }
