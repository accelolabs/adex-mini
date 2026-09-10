package auction

func matches(partner Partner, request Auction) bool {
	return partner.IsEnabled &&
		containsOrAny(partner.Countries, request.Country) &&
		containsOrAny(partner.DeviceTypes, request.DeviceType) &&
		request.BidFloor >= partner.MinBidFloor &&
		!intersects(partner.BlockedCategories, request.Categories)
}

func containsOrAny(values []string, target string) bool {
	if len(values) == 0 {
		return true
	}
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func intersects(left, right []string) bool {
	for _, leftValue := range left {
		for _, rightValue := range right {
			if leftValue == rightValue {
				return true
			}
		}
	}
	return false
}
