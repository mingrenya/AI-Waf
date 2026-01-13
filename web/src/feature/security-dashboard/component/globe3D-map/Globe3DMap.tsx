import React, { useEffect, useRef } from "react"
import { WebGLRenderer, PerspectiveCamera, Scene, AmbientLight, DirectionalLight, Color, PointLight, MeshPhysicalMaterial, DoubleSide } from "three"
import { OrbitControls } from "three/addons/controls/OrbitControls.js"
import ThreeGlobe from "three-globe"
import countries from "./globe-data-min.json"
import type { FeatureCollection, GeoJsonProperties, Geometry } from "./geojson"
import { WAFAttackTrajectory, WAF_ATTACK_TRAJECTORY_COLORS } from "./types"

// WAF安全仪表板全局Globe实例管理 - 避免StrictMode重复创建问题
const wafSecurityGlobeInstance: {
    renderer: WebGLRenderer | null
    camera: PerspectiveCamera | null
    scene: Scene | null
    globe: ThreeGlobe | null
    controls: OrbitControls | null
    animationId: number | null
    isInitialized: boolean
    isCreating: boolean  // 防止并发创建
    elementRef: HTMLDivElement | null
    componentCount: number  // 追踪使用该实例的组件数量
    resizeHandler: (() => void) | null // 存储resize处理函数
} = {
    renderer: null,
    camera: null,
    scene: null,
    globe: null,
    controls: null,
    animationId: null,
    isInitialized: false,
    isCreating: false,
    elementRef: null,
    componentCount: 0,
    resizeHandler: null
}

// WAF安全Globe组件 - 使用全局WebGL上下文管理
const Globe3DMap = React.memo(({ wafAttackTrajectoryData }: { wafAttackTrajectoryData: WAFAttackTrajectory[] }) => {
    const globeRef = useRef<HTMLDivElement>(null)
    const isMountedRef = useRef(false)

    // 初始化或重用全局Three.js实例
    useEffect(() => {
        if (!globeRef.current) return

        const el = globeRef.current
        isMountedRef.current = true
        wafSecurityGlobeInstance.componentCount++

        console.log(`Globe component mounted, count: ${wafSecurityGlobeInstance.componentCount}`)

        // 如果正在创建中，等待创建完成
        if (wafSecurityGlobeInstance.isCreating) {
            console.log('WebGL context is being created, waiting...')
            const checkInterval = setInterval(() => {
                if (!wafSecurityGlobeInstance.isCreating && wafSecurityGlobeInstance.isInitialized) {
                    clearInterval(checkInterval)
                    attachToElement(el)
                }
            }, 50)
            return () => clearInterval(checkInterval)
        }

        // 如果全局实例已存在且有效，重用它
        if (wafSecurityGlobeInstance.isInitialized && wafSecurityGlobeInstance.renderer) {
            console.log('Reusing existing WebGL context...')
            attachToElement(el)
            return
        }

        // 创建新实例
        createGlobeInstance(el)

        // 清理函数
        return () => {
            console.log(`Globe component unmounting, count: ${wafSecurityGlobeInstance.componentCount}`)
            isMountedRef.current = false
            wafSecurityGlobeInstance.componentCount--

            // 只有当没有组件使用时才清理全局实例
            if (wafSecurityGlobeInstance.componentCount <= 0) {
                setTimeout(() => {
                    // 再次检查是否真的没有组件在使用
                    if (wafSecurityGlobeInstance.componentCount <= 0) {
                        console.log('No components using global instance, cleaning up...')
                        cleanupGlobalInstance()
                    }
                }, 1000) // 延迟清理，给路由切换一些时间
            }
        }
    }, [])

    // 将现有canvas附加到新元素
    const attachToElement = (el: HTMLDivElement) => {
        if (!wafSecurityGlobeInstance.renderer) return

        const rect = el.getBoundingClientRect()

        // 如果canvas不在当前容器中，移动它
        if (!el.contains(wafSecurityGlobeInstance.renderer.domElement)) {
            el.innerHTML = ''
            el.appendChild(wafSecurityGlobeInstance.renderer.domElement)
            wafSecurityGlobeInstance.elementRef = el
        }

        // 设置容器样式
        el.style.setProperty("width", "100%")
        el.style.setProperty("height", "100%")
        el.style.setProperty("overflow", "hidden")

        // 更新尺寸
        if (wafSecurityGlobeInstance.camera) {
            wafSecurityGlobeInstance.camera.aspect = rect.width / rect.height
            wafSecurityGlobeInstance.camera.updateProjectionMatrix()
        }
        wafSecurityGlobeInstance.renderer.setSize(rect.width, rect.height)

        // 确保重用实例时也设置正确的旋转角度
        // if (wafSecurityGlobeInstance.globe) {
        //     console.log('Setting rotation for reused globe instance...')
        //     wafSecurityGlobeInstance.globe.rotation.y = -Math.PI * (95 / 180)  // -95度
        //     wafSecurityGlobeInstance.globe.rotation.z = -Math.PI / 8  // 轻微倾斜
        // }
    }

    // 创建新的Globe实例
    const createGlobeInstance = (el: HTMLDivElement) => {
        if (wafSecurityGlobeInstance.isCreating || wafSecurityGlobeInstance.isInitialized) {
            return
        }

        console.log('Creating new global WebGL context...')
        wafSecurityGlobeInstance.isCreating = true

        const rect = el.getBoundingClientRect()
        el.innerHTML = ''

        try {
            // 创建新的Three.js实例
            const renderer = new WebGLRenderer({
                antialias: true,
                logarithmicDepthBuffer: true,
                alpha: true,
                preserveDrawingBuffer: true  // 保持绘制缓冲区
            })
            const camera = new PerspectiveCamera(50, rect.width / rect.height, 0.001, 1000)
            const scene = new Scene()

            renderer.setPixelRatio(window.devicePixelRatio)
            renderer.setSize(rect.width, rect.height)
            el.appendChild(renderer.domElement)
            el.style.setProperty("width", "100%")
            el.style.setProperty("height", "100%")
            el.style.setProperty("overflow", "hidden")

            const controls = new OrbitControls(camera, renderer.domElement)

            // 设置灯光
            const ambientLight = new AmbientLight(new Color(0xFFFFFF), 1)
            const dLight = new DirectionalLight(new Color(0x000000), 0.6)
            dLight.position.set(-400, 100, 400)
            const dLight1 = new DirectionalLight(new Color(0xFFFFFF), 1)
            dLight1.position.set(-200, 500, 200)
            const dLight2 = new PointLight(new Color(0xFFFFFF), 0.8)
            dLight2.position.set(-200, 500, 200)
            const dirLight = new DirectionalLight(0x000000, 1)
            dirLight.position.set(5, 3, 4)
            scene.add(ambientLight, dLight, dLight1, dLight2, dirLight)

            // 设置相机
            camera.aspect = rect.width / rect.height
            camera.position.z = 320  // 从400减少到280，让球体显示更大
            camera.position.x = 0
            camera.position.y = 0
            camera.updateProjectionMatrix()

            // 设置控制器
            controls.enableDamping = true
            controls.enablePan = false
            controls.minDistance = 150  // 从200减少到150
            controls.maxDistance = 400  // 从500减少到400
            controls.rotateSpeed = 0.8
            controls.zoomSpeed = 1
            controls.autoRotate = false
            controls.minPolarAngle = Math.PI / 5
            controls.maxPolarAngle = Math.PI - Math.PI / 5

            // 创建地球
            console.log('Creating globe instance...')
            const globe = new ThreeGlobe({ waitForGlobeReady: true, animateIn: true })
            scene.add(globe)

            const globeMaterial = new MeshPhysicalMaterial({
                color: new Color(0x0d0c27),
                transparent: true,       // 启用透明
                transmission: 1,         // 透光率
                thickness: 0.5,          // 厚度
                roughness: 0.75,         // 粗糙度
                metalness: 0,            // 金属感
                ior: 1.5,                // 折射率
                envMapIntensity: 1.5,    // 环境贴图强度
                reflectivity: 0.1,       // 反射率
                opacity: 0.55,           // 不透明度
                side: DoubleSide,        // 双面渲染
            })

            globe
                .showGlobe(true)
                .globeMaterial(globeMaterial)
                .showAtmosphere(true)
                .atmosphereColor("#0d0c27")
                .atmosphereAltitude(0.25)

            globe
                .hexPolygonsData(countries.features)
                .hexPolygonResolution(3)
                .hexPolygonAltitude(0.001)
                .hexPolygonMargin(0.4)
                .hexPolygonColor((e) => {
                    const countryCode = (e as FeatureCollection<Geometry, GeoJsonProperties>["features"][number]).properties!.ISO_A3
                    return (countryCode === "CHN" || countryCode === "TWN") ? "rgba(255, 255, 255, 1)" : "rgba(241, 230, 255, 1)"
                })

            // 设置 onGlobeReady 回调（只设置一次）
            globe.onGlobeReady(() => {
                console.log('Globe is ready!')
                // 在地球准备好后设置旋转角度
                // globe.rotation.y = -Math.PI * (95 / 180)  // -95度
                // globe.rotation.z = -Math.PI / 8  // 轻微倾斜以获得更好的视角
            })

            // 立即设置旋转角度（不等待onGlobeReady）
            // console.log('Setting globe rotation immediately...')
            // globe.rotation.y = -Math.PI * (95 / 180)  // -95度
            // globe.rotation.z = -Math.PI / 8  // 轻微倾斜以获得更好的视角

            // 更新全局实例
            wafSecurityGlobeInstance.renderer = renderer
            wafSecurityGlobeInstance.camera = camera
            wafSecurityGlobeInstance.scene = scene
            wafSecurityGlobeInstance.globe = globe
            wafSecurityGlobeInstance.controls = controls
            wafSecurityGlobeInstance.elementRef = el
            wafSecurityGlobeInstance.isInitialized = true
            wafSecurityGlobeInstance.isCreating = false

            // 动画循环
            const animate = () => {
                if (!wafSecurityGlobeInstance.isInitialized || !isMountedRef.current) return

                camera.lookAt(scene.position)
                controls.update()
                renderer.render(scene, camera)
                wafSecurityGlobeInstance.animationId = requestAnimationFrame(animate)
            }

            console.log('Starting animation loop...')
            animate()

            // 窗口大小变化处理
            const handleResize = () => {
                if (!wafSecurityGlobeInstance.camera || !wafSecurityGlobeInstance.renderer) return
                if (!wafSecurityGlobeInstance.elementRef) return

                const newRect = wafSecurityGlobeInstance.elementRef.getBoundingClientRect()
                wafSecurityGlobeInstance.camera.aspect = newRect.width / newRect.height
                wafSecurityGlobeInstance.camera.updateProjectionMatrix()
                wafSecurityGlobeInstance.renderer.setSize(newRect.width, newRect.height)
            }

            // 保存 resize 处理函数的引用
            wafSecurityGlobeInstance.resizeHandler = handleResize
            window.addEventListener('resize', handleResize)

        } catch (error) {
            console.error('Failed to create WebGL context:', error)
            wafSecurityGlobeInstance.isCreating = false
        }
    }

    // 清理全局实例
    const cleanupGlobalInstance = () => {
        console.log('Cleaning up global WebGL context...')

        if (wafSecurityGlobeInstance.animationId) {
            cancelAnimationFrame(wafSecurityGlobeInstance.animationId)
        }

        if (wafSecurityGlobeInstance.renderer) {
            wafSecurityGlobeInstance.renderer.dispose()
            wafSecurityGlobeInstance.renderer.forceContextLoss()
        }

        if (wafSecurityGlobeInstance.controls) {
            wafSecurityGlobeInstance.controls.dispose()
        }

        if (wafSecurityGlobeInstance.scene) {
            wafSecurityGlobeInstance.scene.clear()
        }

        // 正确移除事件监听器
        if (wafSecurityGlobeInstance.resizeHandler) {
            window.removeEventListener('resize', wafSecurityGlobeInstance.resizeHandler)
            wafSecurityGlobeInstance.resizeHandler = null
        }

        // 重置全局实例
        wafSecurityGlobeInstance.renderer = null
        wafSecurityGlobeInstance.camera = null
        wafSecurityGlobeInstance.scene = null
        wafSecurityGlobeInstance.globe = null
        wafSecurityGlobeInstance.controls = null
        wafSecurityGlobeInstance.animationId = null
        wafSecurityGlobeInstance.isInitialized = false
        wafSecurityGlobeInstance.isCreating = false
        wafSecurityGlobeInstance.elementRef = null
    }

    // 更新轨迹数据
    useEffect(() => {
        if (!wafSecurityGlobeInstance.globe || !wafSecurityGlobeInstance.isInitialized || !wafAttackTrajectoryData || wafAttackTrajectoryData.length === 0) {
            return
        }

        console.log('Updating globe data, trajectoryData length:', wafAttackTrajectoryData.length)

        const globe = wafSecurityGlobeInstance.globe

        // // 确保每次数据更新时都设置正确的旋转角度
        // console.log('Setting rotation in data update...')
        // // 先重置旋转
        // globe.rotation.set(0, 0, 0)
        // // 然后应用新的旋转
        // globe.rotation.y = -Math.PI * (95 / 180)  // -95度
        // globe.rotation.z = -Math.PI / 8  // 轻微倾斜
        // console.log('Current globe rotation:', globe.rotation.y, globe.rotation.z)

        // 生成WAF攻击源点和防护中心数据
        const attackPointsData = Array.from(wafAttackTrajectoryData
            .reduce((p, c) => {
                // 攻击源点
                p.set(`${c.startLat},${c.startLng}`, {
                    lat: c.startLat,
                    lng: c.startLng,
                })
                // WAF防护中心 (杭州)
                p.set(`${c.endLat},${c.endLng}`, {
                    lat: c.endLat,
                    lng: c.endLng,
                })
                return p
            }, new Map<string, { lat: number; lng: number }>()).values()
        )

        // 更新WAF攻击轨迹数据 - 优化线条样式
        globe
            .arcsData(wafAttackTrajectoryData)
            .arcColor((e: unknown) => {
                const attackTrajectory = e as WAFAttackTrajectory
                // 根据攻击类型分配颜色
                return WAF_ATTACK_TRAJECTORY_COLORS[attackTrajectory.colorIndex]
            })
            .arcAltitude((e: unknown) => Math.pow((e as WAFAttackTrajectory).arcAlt * 0.22, 1.225)) // 攻击轨迹高度(因为此处的arcAlt是地理距离, 所以需要适配四次贝塞尔曲线的高度, 距离越远需要给的补偿值越大因此使用倍率加次方计算)
            .arcStroke(0.3) // 攻击轨迹线条粗细
            .arcDashLength(0.9)
            .arcDashGap(4)
            .arcDashAnimateTime(1000)
            .arcsTransitionDuration(1000)
            .arcDashInitialGap((e: unknown) => (e as WAFAttackTrajectory).order * 1)

        // 使用环形动画效果标记攻击点位
        globe
            .ringsData(attackPointsData)
            .pointColor("#ffffaa")
            .pointsMerge(true)
            .pointRadius(0.25)

        console.log('Globe data updated successfully!')
    }, [wafAttackTrajectoryData])

    return <div className="w-full h-full" ref={globeRef}></div>
})

Globe3DMap.displayName = 'Globe3DMap'

export default Globe3DMap 