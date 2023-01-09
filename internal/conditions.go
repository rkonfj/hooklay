package internal

import (
	"fmt"
	"strings"

	"github.com/oliveagle/jsonpath"
)

type Condition struct {
	Key      string
	Operator string
	Value    string
}

type Conditions struct {
}

func NewConditions() *Conditions {
	return &Conditions{}
}

func (c *Conditions) Meet(rawRequestBody map[string]any, conditions []Condition) error {
	for _, condition := range conditions {
		if strings.EqualFold("eq", condition.Operator) {
			value, err := jsonpath.JsonPathLookup(rawRequestBody, condition.Key)
			if err != nil {
				return err
			}
			if value != condition.Value {
				return fmt.Errorf("Condition failed: %s not meet %s %s",
					value, condition.Operator, condition.Value)
			}
		}
	}
	return nil
}
