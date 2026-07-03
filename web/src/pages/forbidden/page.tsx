import { useNavigate } from 'react-router'
import { ShieldOff } from 'lucide-react'

export default function ForbiddenPage() {
  const navigate = useNavigate()

  return (
    <div className="min-h-screen flex items-center justify-center p-6" style={{ background: 'linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%)' }}>
      <div className="glass-card text-center max-w-md w-full p-10 animate-scale-in">
        <div className="w-20 h-20 mx-auto mb-6 rounded-full flex items-center justify-center"
             style={{ background: 'rgba(239, 68, 68, 0.15)', border: '1px solid rgba(239, 68, 68, 0.3)' }}>
          <ShieldOff className="w-10 h-10 text-red-400" />
        </div>
        <h1 className="text-3xl font-bold text-white mb-3">403 — 无权限访问</h1>
        <p className="text-white/60 mb-8 leading-relaxed">
          您当前的账户角色没有访问此页面的权限。<br />
          如需提升权限，请联系系统管理员。
        </p>
        <div className="flex items-center gap-3 justify-center">
          <button onClick={() => navigate(-1)} className="glass-btn px-6 py-2.5 text-sm">
            返回上一页
          </button>
          <button onClick={() => navigate('/')}
            className="px-6 py-2.5 rounded-xl text-sm font-semibold text-white transition-all duration-300 hover:shadow-lg"
            style={{ background: 'linear-gradient(135deg, #667eea, #764ba2)' }}>
            前往首页
          </button>
        </div>
      </div>
    </div>
  )
}
