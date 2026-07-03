import type { Config } from "tailwindcss"
import tailwindcssAnimate from "tailwindcss-animate"

const config = {
    darkMode: ["class"],
    content: [
        "./pages/**/*.{ts,tsx}",
        "./components/**/*.{ts,tsx}",
        "./app/**/*.{ts,tsx}",
        "./src/**/*.{ts,tsx}",
        "*.{js,ts,jsx,tsx,mdx}",
    ],
    prefix: "",
    theme: {
        extend: {
            container: {
                center: true,
                padding: "2rem",
                screens: {
                    "2xl": "1400px",
                },
            },
            borderRadius: {
                lg: 'var(--radius)',
                md: 'calc(var(--radius) - 2px)',
                sm: 'calc(var(--radius) - 4px)'
            },
            colors: {
                background: 'hsl(var(--background))',
                foreground: 'hsl(var(--foreground))',
                card: {
                    DEFAULT: 'hsl(var(--card))',
                    foreground: 'hsl(var(--card-foreground))'
                },
                popover: {
                    DEFAULT: 'hsl(var(--popover))',
                    foreground: 'hsl(var(--popover-foreground))'
                },
                primary: {
                    DEFAULT: 'hsl(var(--primary))',
                    foreground: 'hsl(var(--primary-foreground))',
                },
                secondary: {
                    DEFAULT: 'hsl(var(--secondary))',
                    foreground: 'hsl(var(--secondary-foreground))'
                },
                muted: {
                    DEFAULT: 'hsl(var(--muted))',
                    foreground: 'hsl(var(--muted-foreground))'
                },
                accent: {
                    DEFAULT: 'hsl(var(--accent))',
                    foreground: 'hsl(var(--accent-foreground))',
                },
                destructive: {
                    DEFAULT: 'hsl(var(--destructive))',
                    foreground: 'hsl(var(--destructive-foreground))'
                },
                border: 'hsl(var(--border))',
                input: 'hsl(var(--input))',
                ring: 'hsl(var(--ring))',
                chart: {
                    '1': 'hsl(var(--chart-1))',
                    '2': 'hsl(var(--chart-2))',
                    '3': 'hsl(var(--chart-3))',
                    '4': 'hsl(var(--chart-4))',
                    '5': 'hsl(var(--chart-5))'
                },

                iconStroke: {
                    light: '##8861DB',      // 亮色背景上使用
                    DEFAULT: '#8861DB',    // 默认颜色
                    dark: '#8861DB',       // 暗色背景上使用
                    accent: '#8861DB',     // 强调色
                },
            },
            keyframes: {
                "accordion-down": {
                    from: { height: "0" },
                    to: { height: "var(--radix-accordion-content-height)" },
                },
                "accordion-up": {
                    from: { height: "var(--radix-accordion-content-height)" },
                    to: { height: "0" },
                },
                "icon-shake": {
                    "0%": { transform: "rotate(0deg)" },
                    "25%": { transform: "rotate(-12deg)" },
                    "50%": { transform: "rotate(10deg)" },
                    "75%": { transform: "rotate(-6deg)" },
                    "85%": { transform: "rotate(3deg)" },
                    "92%": { transform: "rotate(-2deg)" },
                    "100%": { transform: "rotate(0deg)" },
                },
                "pulse-glow": {
                    "0%, 100%": { opacity: "0.6", transform: "scale(1)" },
                    "50%": { opacity: "0.8", transform: "scale(1.1)" },
                },
                "fade-in-up": {
                    "0%": { opacity: "0", transform: "translateY(20px)" },
                    "100%": { opacity: "1", transform: "translateY(0)" },
                },
                "wiggle": {
                    "0%, 100%": { transform: "rotate(-2deg)" },
                    "50%": { transform: "rotate(2deg)" },
                },
                "fade-in": {
                    "0%": { opacity: "0" },
                    "100%": { opacity: "1" },
                },
                "slide-up": {
                    "0%": { opacity: "0", transform: "translateY(16px)" },
                    "100%": { opacity: "1", transform: "translateY(0)" },
                },
                "slide-in-right": {
                    "0%": { opacity: "0", transform: "translateX(16px)" },
                    "100%": { opacity: "1", transform: "translateX(0)" },
                },
                "scale-in": {
                    "0%": { opacity: "0", transform: "scale(0.95)" },
                    "100%": { opacity: "1", transform: "scale(1)" },
                },
                "pulse-green": {
                    "0%, 100%": { opacity: "1" },
                    "50%": { opacity: "0.4" },
                },
                "ping-slow": {
                    "0%": { transform: "scale(1)", opacity: "1" },
                    "100%": { transform: "scale(1.5)", opacity: "0" },
                },
                "gradient-shift": {
                    "0%": { backgroundPosition: "0% 50%" },
                    "50%": { backgroundPosition: "100% 50%" },
                    "100%": { backgroundPosition: "0% 50%" },
                },
            },
            animation: {
                "accordion-down": "accordion-down 0.2s ease-out",
                "accordion-up": "accordion-up 0.2s ease-out",
                "icon-shake": "icon-shake 0.7s ease-out",
                "pulse-glow": "pulse-glow 4s ease-in-out infinite",
                "fade-in-up": "fade-in-up 0.5s ease-out",
                "wiggle": "wiggle 1s ease-in-out infinite",
                "fade-in": "fade-in 0.4s ease-out",
                "slide-up": "slide-up 0.5s ease-out",
                "slide-in-right": "slide-in-right 0.4s ease-out",
                "scale-in": "scale-in 0.3s ease-out",
                "pulse-green": "pulse-green 2s ease-in-out infinite",
                "ping-slow": "ping-slow 2s ease-out infinite",
                "gradient-shift": "gradient-shift 8s ease infinite",
            },
            fontFamily: {
                sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
                mono: ["JetBrains Mono", "ui-monospace", "monospace"],
            },
            backdropBlur: {
                xs: "2px",
                sm: "4px",
                md: "10px",
                lg: "20px",
                xl: "40px",
            },
            boxShadow: {
                "glass": "0 8px 32px 0 rgba(31,38,135,0.37)",
                "glass-hover": "0 12px 48px 0 rgba(31,38,135,0.45)",
                "glass-light": "0 4px 16px 0 rgba(31,38,135,0.2)",
            },
        },
    },
    plugins: [
        tailwindcssAnimate,
    ],
} satisfies Config

export default config
