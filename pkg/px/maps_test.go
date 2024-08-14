package px

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNested2Dotted(t *testing.T) {
	assert := assert.New(t)
	src := map[string]any{
		"top1": 1,
		"top2": "two",
		"p1": map[string]any{
			"c1": 3,
		},
		"p4": map[string]any{
			"c4": map[string]any{
				"cc5": 5,
			},
		},
	}
	new, err := nested2dotted(src)
	assert.NoError(err)
	assert.Equal(1, new["top1"])
	assert.Equal("two", new["top2"])
	assert.Equal(3, new["p1.c1"])
	assert.Equal(5, new["p4.c4.cc5"])
}

func TestNested2DottedError(t *testing.T) {
	assert := assert.New(t)
	src := map[string]any{
		"p4": map[string]any{
			"c4": map[string]any{
				"cc5": map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c": map[string]any{
								"d": map[string]any{
									"e": map[string]any{
										"f": map[string]any{
											"g": map[string]any{
												"h": map[string]any{
													"i": 237,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := nested2dotted(src)
	assert.Error(err)
}
