package config

const (
	// error constants
	CREDENTIALS_ERROR      = "CREDENTIALS_ERROR" // login veya password yanlis
	ACCT_TTL_NOT_VALID     = "ACCT_TTL_NOT_VALID"
	ACCT_GENERATION_ERROR  = "ACCT_GENERATION_ERROR" // uretemedi
	REFT_GENERATION_ERROR  = "REFT_GENERATION_ERROR"
	DBQUERY_ERROR          = "DBQUERY_ERROR"
	REFT_TTL_NOT_VALID     = "REFT_TTL_NOT_VALID" // gecerli degil
	AUTH_REQUIRED          = "AUTH_REQUIRED"      // login etsene demek
	INSERCURE_PWD          = "INSERCURE_PWD"
	ACCT_EXPIRED           = "ACCT_EXPIRED"   // yeniden kimligini dogrula
	SERVER_ERROR           = "SERVER_ERROR"   // server calisamiyor
	REFT_NOT_FOUND         = "REFT_NOT_FOUND" // token bulunamadi
	REFT_EXPIRED           = "REFT_EXPIRED"
	NOT_FOUND              = "NOT_FOUND"     // employee not found mesela
	CANT_DECODE            = "CANT_DECODE"   // decode edemedi
	TYPE_INVALID           = "TYPE_INVALID"  // type dogru degil
	NO_PERMISSION          = "NO_PERMISSION" // eger admin olsaydin yapabilirdin demek gibi
	STATUS_ABORTED         = "ABORTED"
	PACKAGE_EXPIRED        = "PACKAGE_EXPIRED"    // suresi doldu
	NOT_ALLOWED            = "NOT_ALLOWED"        // bu operasyona hic kimsenin izni yok demek gibi
	QUERY_NOT_PROVIDED     = "QUERY_NOT_PROVIDED" // query parametresi verilmemis
	PARAM_NOT_PROVIDED     = "PARAM_NOT_PROVIDED" // param parametresi verilmemis
	BODY_NOT_PROVIDED      = "BODY_NOT_PROVIDED"  // param parametresi verilmemis
	TOO_MANY_REQUESTS      = "TOO_MANY_REQUESTS"
	CANT_UPDATE            = "CANT_UPDATE"
	CANT_INSERT            = "CANT_INSERT"
	CANT_DELETE            = "CANT_DELETE"
	TRANSACTION_FAILED     = "TRANSACTION_FAILED"
	TRANSACTION_SUCCESSFUL = "TRANSACTION_SUCCESSFUL"

	// response strings
	REMOVED = "REMOVED_SUCCESSFULLY" // var olan bir post mesela, listeden cikarildi ama silinmedi var hala
	ADDED   = "ADDED_SUCCESSFULLY"   // mesela bir post listeye eklendi
	UPDATED = "UPDATED_SUCCESSFULLY" // mesela bir post guncellendi
	CREATED = "CREATED_SUCCESSFULLY" // daha onceden olmayan bir post yaratildi mesela
	DELETED = "DELETED_SUCCESSFULLY" // daha onceden var olan bir post databaseden silindi mesela

	// database collections
	EMPLOYEES           = "employees"
	CUSTOMERS           = "customers"
	SELLERS             = "sellers"
	NOTIFICATION_TOKENS = "notification_tokens"
	FAV_SELLERS         = "favorite_sellers"
	PRODUCTS            = "products"
	FAV_PRODS           = "favorite_products"
	POSTS               = "posts"
	FAV_POSTS           = "favorite_posts"
	REASONS             = "reasons"
	NOTES               = "notes"
	SUGGESTIONS         = "suggestions"
	NOTIFICATIONS       = "notifications"
	FEEDBACKS           = "feedbacks"
	EVERYDAYWORK        = "everyday_work"
	CASHIERWORKS        = "cashier_works"
	TRANSFERS           = "transfers"
	PACKAGEHISTORY      = "package_history"
	BRANDS              = "brands"
	BANNERS             = "banners"
	CATEGORIES          = "categories"
	ATTRIBUTES          = "attributes"
	CITIES              = "cities"
	COLLECTIONS         = "collections"
	WALLETS             = "wallets"
	WALLETHISTORY       = "wallet_history"
	PACKAGES            = "packages"
	AUCTIONS            = "auctions"
	CRON_JOBS           = "cron_jobs"
	TASKS               = "tasks"
	// variables
	CURRENT_EMPLOYEE = "currentEmployee"
	CURRENT_USER     = "currentUser"
	EMPLOYEE_JOB     = "employee_job"
	MIN_PWD_ENTROPY  = float64(50)
	ACCT_EXPIREDIN   = "ACCESS_TOKEN_EXPIRED_IN"
	REFT_EXPIREDIN   = "REFRESH_TOKEN_EXPIRED_IN"
	// token keys
	ACCT_PRIVATE_KEY = "ACCESS_TOKEN_PRIVATE_KEY"
	ACCT_PUBLIC_KEY  = "ACCESS_TOKEN_PUBLIC_KEY"
	REFT_PRIVATE_KEY = "REFRESH_TOKEN_PRIVATE_KEY"
	REFT_PUBLIC_KEY  = "REFRESH_TOKEN_PUBLIC_KEY"
	// default images
	DEFAULT_SELLER_IMAGE  = "DEFAULT_SELLER_IMAGE"
	DEFAULT_USER_IMAGE    = "DEFAULT_USER_IMAGE"
	DEFAULT_RBEE_IMAGE    = "DEFAULT_REPORTER_BEE_IMAGE"
	DEFAULT_BANNER_IMAGE  = "DEFAULT_BANNER_IMAGE"
	DEFAULT_AUCTION_IMAGE = "DEFAULT_AUCTION_IMAGE"
	// folders
	FOLDER_SELLERS            = "sellers"
	FOLDER_USERS              = "users"
	FOLDER_POSTS              = "posts"
	FOLDER_PRODUCTS           = "products"
	FOLDER_AUCTIONS           = "auctions"
	FOLDER_BANNERS            = "banners"
	FOLDER_BRANDS             = "brands"
	FOLDER_CATEGORIES         = "categories"
	FOLDER_EMPLOYEE_AVATARS   = "employees/avatars"
	FOLDER_EMPLOYEE_PASSPORTS = "employees/passports"
	// intents
	INTENT_WITHDRAW = "withdraw"
	INTENT_PAYMENT  = "payment"
	INTENT_DEPOSIT  = "deposit"
	// statuses
	STATUS_WAITING   = "waiting"
	STATUS_RECEIVED  = "received"
	STATUS_CANCELLED = "cancelled"
	STATUS_REJECTED  = "rejected"
	STATUS_COMPLETED = "completed"
	STATUS_PUBLISHED = "published"
	STATUS_FINISHED  = "finished"
	STATUS_DELETED   = "deleted"
	STATUS_CHECKING  = "checking"
	STATUS_EXPIRED   = "expired"
	STATUS_ACTIVE    = "active"
	STATUS_CLOSED    = "closed"
	STATUS_BLOCKED   = "blocked"
	// package payment actions
	PACKAGE_PAY    = "pay"
	PACKAGE_CHANGE = "changed"
	// package types = versions
	PACKAGE_TYPE_BASIC    = "basic"
	PACKAGE_TYPE_STANDARD = "standard"
	PACKAGE_TYPE_PREMIUM  = "premium"
	// seller statuses
	SELLER_STATUS_CHECKING  = "checking"
	SELLER_STATUS_PUBLISHED = "published"
	SELLER_STATUS_BLOCKED   = "blocked"
	SELLER_STATUS_NOTPAID   = "notpaid"
	SELLER_STATUS_DELETED   = "deleted"
	// seller types
	SELLER_TYPE_MANUFACTURER = "manufacturer"
	SELLER_TYPE_REGULAR      = "regular"
	SELLER_TYPE_REPORTERBEE  = "reporterbee"
	// task types
	TASK_AUCTION      = "auction"
	TASK_NOTIFICATION = "notification"
	TASK_POST         = "post"
	TASK_PRODUCT      = "product"
	TASK_PROFILE      = "profile"
	// cron job
	PERMISSION_STARTED = "permission_started"
	PERMISSION_ENDED   = "permission_ended"
	// events for cron_jobs service
	AUCTION_FINISHED       = "auction_finished"
	AUCTION_REMOVED        = "auction_removed"
	ADD_DISCOUNT           = "add_discount"
	REMOVE_DISCOUNT        = "remove_discount"
	CANCEL_WITHDRAW_ACTION = "cancel_withdraw_action"
	// discount types
	DISCOUNT_PERCENT = "percent"
	DISCOUNT_PRICE   = "price"
	// duration types
	DURATION_HOUR  = "hour"
	DURATION_DAY   = "day"
	DURATION_MONTH = "month"

	// employee job names, roles
	ADMIN             = "admin"
	EMPLOYEES_MANAGER = "employees_manager"
	ADMIN_CHECKER     = "admin_checker"
	// ACCOUNTANT        = "accountant"
	OWNER   = "owner"
	CASHIER = "cashier"
)

var ALL_EMPLOYEES = []string{ADMIN, EMPLOYEES_MANAGER, ADMIN_CHECKER, OWNER, CASHIER}
var PACKAGE_TYPES = []string{PACKAGE_TYPE_BASIC, PACKAGE_TYPE_STANDARD, PACKAGE_TYPE_PREMIUM}
var PACKAGE_PAY_CHANGE = []string{PACKAGE_PAY, PACKAGE_CHANGE}
var DURATION_TYPES = []string{DURATION_HOUR, DURATION_DAY, DURATION_MONTH}
var DISCOUNT_TYPES = []string{DISCOUNT_PERCENT, DISCOUNT_PRICE}
var WALLET_STATUSES = []string{STATUS_ACTIVE, STATUS_CLOSED}
var TRANSFER_STATUSES = []string{STATUS_CANCELLED, STATUS_COMPLETED, STATUS_EXPIRED, STATUS_WAITING}
var SELLER_STATUSES = []string{SELLER_STATUS_CHECKING, SELLER_STATUS_PUBLISHED, SELLER_STATUS_BLOCKED, SELLER_STATUS_NOTPAID, SELLER_STATUS_DELETED}
var IMAGE_EXTENSIONS = []string{"jpg", "jpeg", "png", "webp"}
