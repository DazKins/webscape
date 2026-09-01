package component

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sort"
	"strings"
	"webscape/server/util"
)

const ComponentIdAppearance = ComponentId("appearance")

var (
	skinTones      = []string{"porcelain", "fair", "tan", "brown", "deep"}
	hairStyles     = []string{"cropped", "swept", "bob", "curls"}
	hairColors     = []string{"black", "darkBrown", "chestnut", "auburn", "golden", "gray"}
	tunicColors    = []string{"slateBlue", "forest", "rust", "mustard", "plum", "teal", "burgundy"}
	trousersColors = []string{"charcoal", "navy", "umber", "olive", "taupe"}
	shoeColors     = []string{"darkBrown", "oxblood", "charcoal", "tan"}
)

type Appearance struct {
	SkinTone      string
	HairStyle     string
	HairColor     string
	TunicColor    string
	TrousersColor string
	ShoeColor     string
}

type CAppearance struct {
	appearance Appearance
}

func NewCAppearance(appearance Appearance) (*CAppearance, error) {
	if err := ValidateAppearance(appearance); err != nil {
		return nil, err
	}
	return &CAppearance{appearance: appearance}, nil
}

func RandomAppearance() Appearance {
	return Appearance{
		SkinTone:      randomChoice(skinTones),
		HairStyle:     randomChoice(hairStyles),
		HairColor:     randomChoice(hairColors),
		TunicColor:    randomChoice(tunicColors),
		TrousersColor: randomChoice(trousersColors),
		ShoeColor:     randomChoice(shoeColors),
	}
}

func DeterministicAppearance(entityID string) Appearance {
	return Appearance{
		SkinTone:      deterministicChoice(entityID, "skinTone", skinTones),
		HairStyle:     deterministicChoice(entityID, "hairStyle", hairStyles),
		HairColor:     deterministicChoice(entityID, "hairColor", hairColors),
		TunicColor:    deterministicChoice(entityID, "tunicColor", tunicColors),
		TrousersColor: deterministicChoice(entityID, "trousersColor", trousersColors),
		ShoeColor:     deterministicChoice(entityID, "shoeColor", shoeColors),
	}
}

func ParseAppearance(value any) (Appearance, error) {
	raw, ok := value.(map[string]any)
	if !ok {
		return Appearance{}, fmt.Errorf("appearance must be an object")
	}
	allowed := map[string]bool{
		"skinTone": true, "hairStyle": true, "hairColor": true,
		"tunicColor": true, "trousersColor": true, "shoeColor": true,
	}
	unknown := make([]string, 0)
	for key := range raw {
		if !allowed[key] {
			unknown = append(unknown, key)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		return Appearance{}, fmt.Errorf("appearance contains unknown field %q", unknown[0])
	}

	appearance := Appearance{}
	var fields = []struct {
		name   string
		target *string
	}{
		{name: "skinTone", target: &appearance.SkinTone},
		{name: "hairStyle", target: &appearance.HairStyle},
		{name: "hairColor", target: &appearance.HairColor},
		{name: "tunicColor", target: &appearance.TunicColor},
		{name: "trousersColor", target: &appearance.TrousersColor},
		{name: "shoeColor", target: &appearance.ShoeColor},
	}
	for _, field := range fields {
		value, ok := raw[field.name].(string)
		if !ok || value == "" {
			return Appearance{}, fmt.Errorf("appearance.%s must be a non-empty string", field.name)
		}
		*field.target = value
	}
	if err := ValidateAppearance(appearance); err != nil {
		return Appearance{}, err
	}
	return appearance, nil
}

func ValidateAppearance(appearance Appearance) error {
	checks := []struct {
		name    string
		value   string
		allowed []string
	}{
		{name: "skinTone", value: appearance.SkinTone, allowed: skinTones},
		{name: "hairStyle", value: appearance.HairStyle, allowed: hairStyles},
		{name: "hairColor", value: appearance.HairColor, allowed: hairColors},
		{name: "tunicColor", value: appearance.TunicColor, allowed: tunicColors},
		{name: "trousersColor", value: appearance.TrousersColor, allowed: trousersColors},
		{name: "shoeColor", value: appearance.ShoeColor, allowed: shoeColors},
	}
	for _, check := range checks {
		if !contains(check.allowed, check.value) {
			return fmt.Errorf("appearance.%s must be one of %s", check.name, strings.Join(check.allowed, ", "))
		}
	}
	return nil
}

func (c *CAppearance) GetId() ComponentId {
	return ComponentIdAppearance
}

func (c *CAppearance) Serialize() util.Json {
	return util.JObject{
		"skinTone":      util.JString(c.appearance.SkinTone),
		"hairStyle":     util.JString(c.appearance.HairStyle),
		"hairColor":     util.JString(c.appearance.HairColor),
		"tunicColor":    util.JString(c.appearance.TunicColor),
		"trousersColor": util.JString(c.appearance.TrousersColor),
		"shoeColor":     util.JString(c.appearance.ShoeColor),
	}
}

func (c *CAppearance) GetAppearance() Appearance {
	return c.appearance
}

func randomChoice(values []string) string {
	return values[rand.Intn(len(values))]
}

func deterministicChoice(entityID string, category string, values []string) string {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(category))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(entityID))
	return values[hash.Sum64()%uint64(len(values))]
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
