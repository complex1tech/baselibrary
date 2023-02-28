package basic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTimeRandomBin64(t *testing.T) {
	u0 := TimeRandomBin64().String()
	u1 := TimeRandomBin64().String()
	assert.NotEqual(t, u0, u1)
}

func TestBin64Pattern__should_match_string(t *testing.T) {
	s0 := (Bin64{}).String()
	s1 := RandomBin64().String()
	s2 := " 341a7d60bc5893a6 "
	s3 := "341a7d60-bc5893a6"

	m0 := Bin64Pattern.MatchString(s0)
	m1 := Bin64Pattern.MatchString(s1)
	m2 := Bin64Pattern.MatchString(s2)
	m3 := Bin64Pattern.MatchString(s3)

	assert.True(t, m0)
	assert.True(t, m1)
	assert.False(t, m2)
	assert.False(t, m3)
}