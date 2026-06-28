const warmNeutral = {
  50: '#fafaf7',
  100: '#f5f1e8',
  200: '#e7ddcc',
  300: '#d4c5ad',
  400: '#aa9b84',
  500: '#766d61',
  600: '#5f574d',
  700: '#453f37',
  800: '#2c2923',
  900: '#1a1814',
  950: '#0a0a0a'
}

const brandGold = {
  50: '#fbf7ed',
  100: '#f5ead2',
  200: '#ead5a4',
  300: '#d6af6e',
  400: '#cda047',
  500: '#bc8a35',
  600: '#a06f28',
  700: '#815722',
  800: '#5f4019',
  900: '#3d2a12',
  950: '#201409'
}

const brandRust = {
  50: '#fff1ee',
  100: '#ffe1dc',
  200: '#ffc7bd',
  300: '#ffa394',
  400: '#f07361',
  500: '#b8413a',
  600: '#9f342f',
  700: '#812a27',
  800: '#672521',
  900: '#461917',
  950: '#260b0a'
}

const mutedEmerald = {
  50: '#eefbf5',
  100: '#d7f4e6',
  200: '#afe8cf',
  300: '#7fd6b2',
  400: '#4dbb91',
  500: '#2f9d75',
  600: '#237e60',
  700: '#1e644f',
  800: '#1b503f',
  900: '#173f34',
  950: '#09251d'
}

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        white: '#fffdf8',
        black: '#0a0a0a',
        gray: warmNeutral,
        slate: warmNeutral,
        zinc: warmNeutral,
        neutral: warmNeutral,
        stone: warmNeutral,
        dark: warmNeutral,
        primary: brandGold,
        accent: brandRust,
        blue: brandGold,
        sky: brandGold,
        cyan: brandGold,
        teal: brandGold,
        amber: brandGold,
        yellow: brandGold,
        orange: brandGold,
        indigo: brandRust,
        violet: brandRust,
        purple: brandRust,
        fuchsia: brandRust,
        pink: brandRust,
        rose: brandRust,
        red: brandRust,
        emerald: mutedEmerald,
        green: mutedEmerald
      },
      fontFamily: {
        sans: [
          'PingFang SC',
          'MiSans',
          'Inter',
          'system-ui',
          '-apple-system',
          'Segoe UI',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'Helvetica Neue',
          'Arial',
          'sans-serif'
        ],
        display: [
          'MiSans',
          'PingFang SC',
          'Inter',
          'system-ui',
          '-apple-system',
          'Segoe UI',
          'sans-serif'
        ],
        mono: ['JetBrains Mono', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 8px 32px rgba(0, 0, 0, 0.08)',
        'glass-sm': '0 4px 16px rgba(0, 0, 0, 0.06)',
        glow: '0 0 20px rgba(188, 138, 53, 0.22)',
        'glow-lg': '0 0 40px rgba(188, 138, 53, 0.32)',
        card: '0 1px 3px rgba(0, 0, 0, 0.04), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'card-hover': '0 10px 40px rgba(0, 0, 0, 0.08)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': 'linear-gradient(135deg, #bc8a35 0%, #815722 100%)',
        'gradient-dark': 'linear-gradient(135deg, #1a1814 0%, #0a0a0a 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 40% 20%, rgba(188, 138, 53, 0.12) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(184, 65, 58, 0.08) 0px, transparent 50%), radial-gradient(at 0% 50%, rgba(188, 138, 53, 0.08) 0px, transparent 50%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(188, 138, 53, 0.22)' },
          '100%': { boxShadow: '0 0 30px rgba(188, 138, 53, 0.34)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
