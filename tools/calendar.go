package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	ics "github.com/arran4/golang-ical"
)

func NewCalendar(icsURL string) *Calendar {
	return &Calendar{
		icsURL: icsURL,
	}
}

type Calendar struct {
	icsURL string
}

type CalendarInput struct {
	Year  int `json:"year" jsonschema_description:"The year to get events for"`
	Month int `json:"month" jsonschema_description:"The month to get events for"`
	Day   int `json:"day" jsonschema_description:"The day to get events for"`
}

func (c *Calendar) Run(ctx context.Context, arguments json.RawMessage) (string, error) {
	fmt.Println("Looking at the calendar to see what's happening today.")
	var input CalendarInput
	if err := json.Unmarshal(arguments, &input); err != nil {
		return "", &InvalidToolArgumentsError{
			ToolName: c.Name(),
			Message:  "invalid JSON format",
		}
	}

	// Validate the date components
	if input.Year < 1900 || input.Year > 2100 {
		return "", &InvalidToolArgumentsError{
			ToolName: c.Name(),
			Message:  "year must be between 1900 and 2100",
		}
	}
	if input.Month < 1 || input.Month > 12 {
		return "", &InvalidToolArgumentsError{
			ToolName: c.Name(),
			Message:  "month must be between 1 and 12",
		}
	}
	if input.Day < 1 || input.Day > 31 {
		return "", &InvalidToolArgumentsError{
			ToolName: c.Name(),
			Message:  "day must be between 1 and 31",
		}
	}

	// Parse the ICS data
	cal, err := ics.ParseCalendarFromUrl(c.icsURL, ctx)
	if err != nil {
		return "", fmt.Errorf("error parsing calendar data: %v", err)
	}

	// Create a new calendar for filtered events
	filteredCal := ics.NewCalendar()
	filteredCal.SetMethod(ics.MethodRequest)
	filteredCal.SetProductId("-//Good Morning//Calendar Tool//EN")
	filteredCal.SetVersion("2.0")

	// Target date for filtering
	targetDate := time.Date(input.Year, time.Month(input.Month), input.Day, 0, 0, 0, 0, time.UTC)

	// Filter events for the specified date
	for _, event := range cal.Events() {
		startTime, err := event.GetStartAt()
		if err != nil {
			continue
		}

		// Check if event occurs on the target date
		if startTime.Year() == targetDate.Year() &&
			startTime.Month() == targetDate.Month() &&
			startTime.Day() == targetDate.Day() {
			// Add the event to the filtered calendar
			filteredCal.AddEvent(event.Id())
			// Copy all properties from the original event
			for _, prop := range event.Properties {
				filteredCal.Events()[0].AddProperty(ics.ComponentProperty(prop.IANAToken), prop.Value)
			}
		}
	}
	fmt.Println("I can see you have " + strconv.Itoa(len(filteredCal.Events())) + " meetings today.")
	// Serialize the filtered calendar
	return filteredCal.Serialize(), nil
}

func (c *Calendar) Name() string {
	return "calendar"
}

func (c *Calendar) ToolDefinition() *anthropic.ToolParam {
	return &anthropic.ToolParam{
		Name:        c.Name(),
		Description: anthropic.String("Get events from the calendar"),
		InputSchema: GenerateSchema[CalendarInput](),
	}
}
