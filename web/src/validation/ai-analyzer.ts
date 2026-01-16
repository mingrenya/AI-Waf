import { z } from "zod"

// AI分析器配置表单验证
export const aiAnalyzerConfigSchema = z.object({
  enabled: z.boolean({
    required_error: "请选择是否启用AI分析",
  }),
  patternDetection: z.object({
    enabled: z.boolean(),
    minSamples: z
      .number({
        required_error: "请输入最小样本数",
        invalid_type_error: "最小样本数必须是数字",
      })
      .min(10, "最小样本数至少为10")
      .max(100000, "最小样本数不能超过100000"),
    anomalyThreshold: z
      .number({
        required_error: "请输入异常阈值",
        invalid_type_error: "异常阈值必须是数字",
      })
      .min(1.0, "异常阈值最小为1.0")
      .max(5.0, "异常阈值最大为5.0"),
    clusteringMethod: z.string(),
    timeWindow: z.number().min(1).max(168), // 1-168小时
  }),
  ruleGeneration: z.object({
    enabled: z.boolean(),
    confidenceThreshold: z
      .number({
        required_error: "请输入置信度阈值",
        invalid_type_error: "置信度阈值必须是数字",
      })
      .min(0.5, "置信度阈值最小为0.5")
      .max(1.0, "置信度阈值最大为1.0"),
    autoDeploy: z.boolean({
      required_error: "请选择是否自动部署",
    }),
    reviewRequired: z.boolean(),
    defaultAction: z.string(),
  }),
  analysisInterval: z
    .number({
      required_error: "请输入分析间隔",
      invalid_type_error: "分析间隔必须是数字",
    })
    .min(5, "分析间隔最少5分钟")
    .max(1440, "分析间隔最多1440分钟（24小时）"),
})

export type AIAnalyzerConfigFormData = z.infer<typeof aiAnalyzerConfigSchema>

// 规则审核表单验证
export const ruleReviewSchema = z.object({
  status: z.enum(["approved", "rejected"], {
    required_error: "请选择审核状态",
  }),
  comment: z.string().max(500, "审核意见最多500字符").optional(),
})

export type RuleReviewFormData = z.infer<typeof ruleReviewSchema>
