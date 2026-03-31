export interface MonthlyTimelineEntry {
  month: string
  totalAmount: string
  billCount: number
}

export interface CategoryBreakdownEntry {
  month: string
  billTypeName: string
  totalAmount: string
  billCount: number
}

export interface MonthlyComplianceEntry {
  month: string
  totalBills: number
  paidOnTime: number
  overdue: number
  complianceRate: string
}

export interface TimelineResponse {
  projectId: string
  months: number
  timeline: MonthlyTimelineEntry[]
}

export interface CategoriesResponse {
  projectId: string
  months: number
  categories: CategoryBreakdownEntry[]
}

export interface ComplianceResponse {
  projectId: string
  months: number
  compliance: MonthlyComplianceEntry[]
}
