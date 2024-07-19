package cerberus

type RuleType int

const (
	IpRule RuleType = iota
	AsnRule
)

type WindowOpts struct {
	Size uint16
	Tick uint16
}

type TriggerOpts interface {
	NewTrigger() Trigger
}

type RuleOpts struct {
	Name    string
	Type    RuleType
	Window  WindowOpts
	Trigger TriggerOpts
	Ignored []string
}

type Options struct {
	Rules []RuleOpts
}

type Rule struct {
	Type    RuleType
	Window  *Window
	Ignored []string
}

type Cerberus struct {
	rules []*Rule
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Analyze is the function that will analyze the metrics and apply the rules
func (c Cerberus) Analyze(ip string, asn string, isLogin bool, isUnauthorized bool) {
	for _, rule := range c.rules {
		if rule.Type == IpRule {
			if contains(rule.Ignored, ip) {
				// If the IP is ignored, we skip the rule
				continue
			}
			// Increment IP metrics
			rule.Window.Inc(ip, isLogin, isUnauthorized)
		} else {
			if contains(rule.Ignored, asn) {
				// If the ASN is ignored, we skip the rule
				continue
			}
			// Increment ASN metrics
			rule.Window.Inc(asn, isLogin, isUnauthorized)
		}
	}
}

func (c Cerberus) Start() {
	// Start the metrics
	for _, rule := range c.rules {
		rule.Window.Start()
	}
}

func NewCerberus(options *Options) *Cerberus {
	rules := make([]*Rule, len(options.Rules))
	for _, ruleOption := range options.Rules {
		trigger := ruleOption.Trigger.NewTrigger()
		window := NewWindow(ruleOption.Name, ruleOption.Window.Tick, ruleOption.Window.Size, trigger)
		rule := &Rule{
			Type:    ruleOption.Type,
			Window:  window,
			Ignored: ruleOption.Ignored,
		}
		rules = append(rules, rule)
	}

	return &Cerberus{rules: rules}
}
