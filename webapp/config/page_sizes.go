package config

// Number of page buttons to show on a pager. Default shows page buttons
// 1 thru N (e.g., 1 thru 8) or w/ your page number in the middle surrounded
// by its neighboring pages.
const (
	PagerButtonLimit = 6 // only even numbers make a difference
)

// Pagination sizes per page.
var (
	PageSizeMemberSearch           = 60
	PageSizeFriends                = 12
	PageSizeBlockList              = 12
	PageSizePrivatePhotoGrantees   = 12
	PageSizeAdminCertification     = 20
	PageSizeAdminFeedback          = 20
	PageSizeSiteGallery            = 16
	PageSizeUserGallery            = 16
	PageSizeInboxList              = 20  // sidebar list
	PageSizeInboxThread            = 10  // conversation view
	PageSizeForums                 = 100 // TODO: for main category index view
	PageSizeThreadList             = 20  // 20 threads per board, 20 posts per thread
	PageSizeForumAdmin             = 20
	PageSizeDashboardNotifications = 50
)
