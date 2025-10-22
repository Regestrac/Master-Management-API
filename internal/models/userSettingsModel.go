package models

import "gorm.io/gorm"

type UserSettings struct {
	gorm.Model
	ID             uint   `json:"id" gorm:"primaryKey"`
	UserId         uint   `json:"user_id"`
	DateFormat     string `json:"date_format"`
	TimeFormat     string `json:"time_format"`
	FirstDayOfWeek string `json:"first_day_of_week"`
	WorkWeek       string `json:"work_week"`
	// DefaultPage    string `json:"default_page"`
	// AutoStartTimer bool   `json:"auto_start_timer"`

	Theme       string `json:"theme"`
	AccentColor string `json:"accent_color"`
	// FontFamily   string `json:"font_family"`
	// FontSize     uint   `json:"font_size"`
	// CompactMode  bool   `json:"compact_mode"`
	// ShowSidebar  bool   `json:"show_sidebar"`
	// SidebarWidth uint   `json:"sidebar_width"`

	FocusDuration     uint `json:"focus_duration"`
	ShortBreak        uint `json:"short_break"`
	LongBreak         uint `json:"long_break"`
	AutoBreak         bool `json:"auto_break"`
	LongBreakAfter    uint `json:"long_break_after"`
	GoalDuration      uint `json:"goal_duration"`
	WeeklyTargetHours uint `json:"weekly_target_hours"`
	// SmartScheduling   bool `json:"smart_scheduling"`
	// AutoCategorize    bool `json:"auto_categorize"`
	// Insights          bool `json:"insights"`
	// TimeBlocking      bool `json:"time_blocking"`

	TaskReminder  bool `json:"task_reminder"`
	GoalProgress  bool `json:"goal_progress"`
	SessionBreaks bool `json:"session_breaks"`
	DailySummary  bool `json:"daily_summary"`
	Milestone     bool `json:"milestone"`
	NewFeature    bool `json:"new_feature"`

	CloudSync             bool   `json:"cloud_sync"`
	KeepCompletedFor      string `json:"keep_completed_for"`
	AnalyticDataRetention string `json:"analytic_data_retention"`
	AutoDeleteOldData     bool   `json:"auto_delete_old_data"`

	// UsageData    bool `json:"usage_data"`
	// Marketing    bool `json:"marketing"`
	// Integrations bool `json:"integrations"`

	DebugMode         bool `json:"debug_mode"`
	BetaFeatures      bool `json:"beta_features"`
	Telemetry         bool `json:"telemetry"`
	AIAssistant       bool `json:"ai_assistant"`
	AdvancedAnalytics bool `json:"advanced_analytics"`
	TeamCollaboration bool `json:"team_collaboration"`
}
