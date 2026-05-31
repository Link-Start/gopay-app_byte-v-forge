package paymentsvc

func extractRedirectToURL(payload map[string]any) string {
	if value := redirectToURLFromAction(objectAt(payload, "next_action")); value != "" {
		return value
	}
	if value := redirectToURLFromIntent(objectAt(payload, "setup_intent")); value != "" {
		return value
	}
	if value := redirectToURLFromIntent(objectAt(payload, "payment_intent")); value != "" {
		return value
	}
	if value := redirectToURLFromIntent(objectAt(payload, "invoice", "payment_intent")); value != "" {
		return value
	}
	return ""
}

func redirectToURLFromIntent(intent map[string]any) string {
	return redirectToURLFromAction(objectAt(intent, "next_action"))
}

func redirectToURLFromAction(action map[string]any) string {
	if stringAt(action, "type") != "redirect_to_url" {
		return ""
	}
	return stringAt(action, "redirect_to_url", "url")
}

func objectAt(value map[string]any, path ...string) map[string]any {
	current := value
	for _, key := range path {
		if current == nil {
			return nil
		}
		next, _ := current[key].(map[string]any)
		current = next
	}
	return current
}
