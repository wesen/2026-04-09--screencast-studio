import { useState, useEffect, useRef, useCallback } from "react";

const C = {
  cream: "#f5f0e8", black: "#1a1a1a", dark: "#2c2c2c", mid: "#8a8a7a",
  light: "#d4d0c8", red: "#c04040", green: "#5a8a5a", amber: "#b89840", bg: "#e8e4dc"
};
const F = `"Chicago","Geneva","Monaco",monospace`;

// ── Primitives ──
const Radio = ({ on }) => (
  <span style={{ display:"inline-block",width:12,height:12,border:`1.5px solid ${C.black}`,borderRadius:6,background:C.cream,marginRight:4,verticalAlign:"middle",position:"relative",flexShrink:0 }}>
    {on && <span style={{ position:"absolute",top:2.5,left:2.5,width:5,height:5,borderRadius:3,background:C.black }} />}
  </span>
);

const Sel = ({ value, opts, onChange, w = 130 }) => (
  <div onClick={() => { const i = opts.indexOf(value); onChange(opts[(i+1)%opts.length]); }}
    style={{ display:"inline-flex",alignItems:"center",background:C.cream,border:`1.5px solid ${C.black}`,borderRadius:2,padding:"1px 3px 1px 6px",fontFamily:F,fontSize:10,cursor:"pointer",width:w,justifyContent:"space-between",height:18 }}>
    <span style={{ overflow:"hidden",textOverflow:"ellipsis",whiteSpace:"nowrap" }}>{value}</span>
    <span style={{ fontSize:7,marginLeft:4 }}>▼</span>
  </div>
);

const Btn = ({ children, active, accent, onClick, disabled, style = {} }) => (
  <button onClick={onClick} disabled={disabled} style={{
    fontFamily:F,fontSize:10,border:`1.5px solid ${C.black}`,borderRadius:3,padding:"3px 10px",cursor:disabled?"default":"pointer",
    background:active?C.black:accent||C.cream,color:active?C.cream:accent?C.cream:C.black,
    boxShadow:active?"none":`1px 1px 0 ${C.dark}`,opacity:disabled?0.4:1,...style
  }}>{children}</button>
);

const Slider = ({ value, onChange, min=0, max=100 }) => {
  const ref = useRef(null);
  const drag = useCallback((e) => {
    const r = ref.current.getBoundingClientRect();
    const x = (e.touches?e.touches[0].clientX:e.clientX);
    onChange(Math.max(min,Math.min(max,Math.round(min+((x-r.left)/r.width)*(max-min)))));
  },[min,max,onChange]);
  const pct = ((value-min)/(max-min))*100;
  return (
    <div ref={ref} onMouseDown={e=>{drag(e);const m=v=>drag(v),u=()=>{window.removeEventListener("mousemove",m);window.removeEventListener("mouseup",u)};window.addEventListener("mousemove",m);window.addEventListener("mouseup",u)}}
      style={{ width:"100%",height:14,cursor:"pointer",position:"relative",display:"flex",alignItems:"center" }}>
      <div style={{ width:"100%",height:3,background:C.light,border:`1px solid ${C.mid}`,borderRadius:1,position:"relative" }}>
        <div style={{ width:`${pct}%`,height:"100%",background:C.mid,borderRadius:1 }} />
      </div>
      <div style={{ position:"absolute",left:`calc(${pct}% - 5px)`,top:1,width:10,height:12,background:C.cream,border:`1.5px solid ${C.black}`,borderRadius:2 }} />
    </div>
  );
};

const WinBar = ({ children, onClose }) => (
  <div style={{ background:`repeating-linear-gradient(to right,${C.black} 0px,${C.black} 2px,${C.cream} 2px,${C.cream} 4px)`,padding:"2px 6px",display:"flex",alignItems:"center",justifyContent:"space-between",borderBottom:`1.5px solid ${C.black}`,height:18,cursor:"grab" }}>
    {onClose && <div onClick={onClose} style={{ width:11,height:11,border:`1.5px solid ${C.black}`,background:C.cream,cursor:"pointer",flexShrink:0,display:"flex",alignItems:"center",justifyContent:"center",fontSize:7,lineHeight:1 }}>✕</div>}
    <div style={{ fontFamily:F,fontSize:10,fontWeight:"bold",color:C.black,background:C.cream,padding:"0 8px",whiteSpace:"nowrap" }}>{children}</div>
    <div style={{ width:11 }} />
  </div>
);

const Win = ({ children, title, onClose, style={} }) => (
  <div style={{ border:`2px solid ${C.black}`,borderRadius:3,background:C.cream,boxShadow:`2px 2px 0 ${C.dark}`,overflow:"hidden",...style }}>
    <WinBar onClose={onClose}>{title}</WinBar>
    <div style={{ padding:6 }}>{children}</div>
  </div>
);

// ── Source Visuals ──
const srcTypes = {
  "Display": { icon: "🖥", scenes: ["Desktop 1","Desktop 2"] },
  "Window":  { icon: "☐", scenes: ["Finder","Terminal","Browser","Code Editor"] },
  "Region":  { icon: "⊞", scenes: ["Top Half","Bottom Half","Custom Region"] },
  "Camera":  { icon: "◉", scenes: ["Built-in","USB Camera","FaceTime HD"] },
};

const FakeScreen = ({ kind, scene }) => {
  if (kind === "Camera") return (
    <div style={{ width:"100%",aspectRatio:"4/3",background:"#1e1e1c",borderRadius:2,border:`1.5px solid ${C.dark}`,display:"flex",alignItems:"center",justifyContent:"center",position:"relative",overflow:"hidden" }}>
      <svg width="40" height="50" viewBox="0 0 40 50"><ellipse cx="20" cy="14" rx="9" ry="10" fill="none" stroke="#606058" strokeWidth="1.2"/><circle cx="16" cy="12" r="1.2" fill="#707068"/><circle cx="24" cy="12" r="1.2" fill="#707068"/><path d="M16 18Q20 22 24 18" fill="none" stroke="#707068" strokeWidth="0.8"/><path d="M8 28Q6 38 5 48L35 48Q34 38 32 28Q28 25 20 25Q12 25 8 28Z" fill="none" stroke="#606058" strokeWidth="1.2"/></svg>
      <div style={{ position:"absolute",inset:0,background:`repeating-linear-gradient(to bottom,transparent 0px,transparent 3px,rgba(0,0,0,0.06) 3px,rgba(0,0,0,0.06) 4px)` }}/>
      <div style={{ position:"absolute",bottom:3,fontFamily:F,fontSize:7,color:"#505048" }}>{scene}</div>
    </div>
  );
  if (kind === "Region") return (
    <div style={{ width:"100%",aspectRatio:"4/3",background:C.black,borderRadius:2,border:`1.5px solid ${C.dark}`,position:"relative",overflow:"hidden" }}>
      <div style={{ position:"absolute",inset:0,background:`repeating-linear-gradient(to bottom,transparent 0,transparent 2px,rgba(245,240,232,0.03) 2px,rgba(245,240,232,0.03) 4px)` }}/>
      <div style={{ position:"absolute",top:"15%",left:"10%",width:"80%",height:"70%",border:`1px dashed ${C.amber}`,borderRadius:1 }}/>
      <div style={{ position:"absolute",top:"15%",left:"10%",fontFamily:F,fontSize:7,color:C.amber,padding:"1px 3px" }}>{scene}</div>
      {["top-left","top-right","bottom-left","bottom-right"].map(c=>{
        const [v,h]=c.split("-"); return <div key={c} style={{ position:"absolute",[v]:v==="top"?"calc(15% - 3px)":"calc(85% - 3px)",[h]:h==="left"?"calc(10% - 3px)":"calc(90% - 3px)",width:6,height:6,border:`1px solid ${C.amber}`,background:"rgba(184,152,64,0.3)" }}/>;
      })}
    </div>
  );
  // Display / Window
  return (
    <div style={{ width:"100%",aspectRatio:"4/3",background:C.black,borderRadius:2,border:`1.5px solid ${C.dark}`,position:"relative",overflow:"hidden" }}>
      <div style={{ position:"absolute",inset:0,background:`repeating-linear-gradient(to bottom,transparent 0,transparent 2px,rgba(245,240,232,0.03) 2px,rgba(245,240,232,0.03) 4px)` }}/>
      <div style={{ position:"absolute",top:5,left:6,right:6 }}>
        <div style={{ fontFamily:F,fontSize:8,color:"#909088",display:"flex",justifyContent:"space-between",marginBottom:3 }}>
          <span>■ {kind==="Window"?scene:"Finder"}</span><span style={{ fontSize:7 }}>12:00</span>
        </div>
        <div style={{ border:`1px solid #505048`,borderRadius:1,padding:3 }}>
          <div style={{ display:"flex",gap:6,flexWrap:"wrap" }}>
            {(kind==="Window"?["src","out","lib"]:["System","Apps","Docs"]).map(f=>(
              <div key={f} style={{ display:"flex",flexDirection:"column",alignItems:"center" }}>
                <div style={{ width:12,height:9,border:`1px solid #606058`,borderRadius:1,marginBottom:1 }}/>
                <span style={{ fontFamily:F,fontSize:6,color:"#707068" }}>{f}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
      <div style={{ position:"absolute",bottom:3,left:0,right:0,textAlign:"center",fontFamily:F,fontSize:7,color:"#505048" }}>{scene||kind}</div>
    </div>
  );
};

const MicMeter = ({ level }) => (
  <div style={{ display:"flex",gap:1,height:10 }}>
    {Array.from({length:16}).map((_,i)=>{
      const p=i/16, on=p<level;
      return <div key={i} style={{ width:4,height:9,background:on?(p>0.75?C.red:p>0.55?C.amber:C.green):C.light,border:`0.5px solid ${on?C.dark:C.mid}`,borderRadius:0.5 }}/>;
    })}
  </div>
);

const Waveform = ({ active }) => {
  const [bars,setBars]=useState(Array(24).fill(2));
  useEffect(()=>{
    if(!active){setBars(Array(24).fill(2));return;}
    const id=setInterval(()=>setBars(p=>p.map(()=>2+Math.random()*14)),90);
    return ()=>clearInterval(id);
  },[active]);
  return (
    <div style={{ display:"flex",alignItems:"flex-end",gap:1,height:18 }}>
      {bars.map((h,i)=><div key={i} style={{ width:2,height:h,background:active?C.green:C.mid,borderRadius:0.5,transition:"height 0.08s" }}/>)}
    </div>
  );
};

// ── App ──
let nextId = 1;
const makeSrc = (kind="Display") => ({
  id: nextId++, kind, scene: srcTypes[kind].scenes[0],
  armed: true, solo: false, label: `${kind} ${nextId-1}`
});

export default function App() {
  const [sources, setSources] = useState([makeSrc("Display"), makeSrc("Camera")]);
  const [recording, setRecording] = useState(false);
  const [paused, setPaused] = useState(false);
  const [elapsed, setElapsed] = useState(0);
  const [format, setFormat] = useState("MOV");
  const [fps, setFps] = useState("24 fps");
  const [quality, setQuality] = useState(75);
  const [audio, setAudio] = useState("48 kHz, 16-bit");
  const [gain, setGain] = useState(55);
  const [micInput, setMicInput] = useState("Built-in Mic");
  const [micLevel, setMicLevel] = useState(0.25);
  const [addOpen, setAddOpen] = useState(false);
  const [multiTrack, setMultiTrack] = useState(true);

  useEffect(()=>{
    if(!recording||paused) return;
    const id=setInterval(()=>setElapsed(e=>e+1),1000);
    return ()=>clearInterval(id);
  },[recording,paused]);

  useEffect(()=>{
    if(!recording||paused){setMicLevel(0.12);return;}
    const id=setInterval(()=>setMicLevel(0.15+Math.random()*0.6),110);
    return ()=>clearInterval(id);
  },[recording,paused]);

  const fmt=s=>`${String(Math.floor(s/3600)).padStart(2,"0")}:${String(Math.floor((s%3600)/60)).padStart(2,"0")}:${String(s%60).padStart(2,"0")}`;

  const updateSrc=(id,patch)=>setSources(s=>s.map(x=>x.id===id?{...x,...patch}:x));
  const removeSrc=id=>setSources(s=>s.filter(x=>x.id!==id));
  const addSrc=kind=>{setSources(s=>[...s,makeSrc(kind)]);setAddOpen(false);};

  const armed = sources.filter(s=>s.armed);
  const diskPct = Math.min(95, 8 + (recording ? elapsed * 0.2 * armed.length : 0));

  return (
    <div style={{ minHeight:"100vh",background:`repeating-linear-gradient(45deg,${C.bg} 0px,${C.bg} 1px,${C.cream} 1px,${C.cream} 6px)`,fontFamily:F,fontSize:11,color:C.black }}>

      {/* Menu */}
      <div style={{ background:C.cream,borderBottom:`2px solid ${C.black}`,padding:"3px 10px",display:"flex",alignItems:"center",gap:14,position:"sticky",top:0,zIndex:20 }}>
        <span style={{ fontSize:14,fontWeight:"bold" }}>⌘</span>
        {["File","Edit","Capture","Sources","Help"].map(m=>(
          <span key={m} style={{ cursor:"pointer",padding:"1px 4px",fontSize:11 }}
            onMouseEnter={e=>{e.target.style.background=C.black;e.target.style.color=C.cream}}
            onMouseLeave={e=>{e.target.style.background="transparent";e.target.style.color=C.black}}>{m}</span>
        ))}
        <div style={{ flex:1 }}/>
        <span style={{ fontSize:9,color:C.mid }}>{armed.length} source{armed.length!==1?"s":""} armed</span>
        {recording&&!paused&&<span style={{ color:C.red,fontWeight:"bold",animation:"blink 1s steps(1) infinite",fontSize:10 }}>● REC</span>}
        <span style={{ fontSize:10,color:C.mid }}>{new Date().toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"})}</span>
      </div>

      <div style={{ padding:"8px 10px",maxWidth:900,margin:"0 auto",display:"flex",flexDirection:"column",gap:10 }}>

        {/* Sources Grid */}
        <Win title={`Sources (${sources.length})`} style={{}}>
          <div style={{ display:"grid",gridTemplateColumns:"repeat(auto-fill, minmax(175px, 1fr))",gap:8 }}>
            {sources.map(src=>(
              <div key={src.id} style={{
                border:`1.5px solid ${src.armed?(recording&&!paused?C.red:C.black):C.mid}`,
                borderRadius:3,background:C.cream,overflow:"hidden",
                opacity:src.armed?1:0.55,transition:"opacity 0.2s"
              }}>
                <div style={{
                  display:"flex",alignItems:"center",justifyContent:"space-between",padding:"2px 5px",
                  background:src.armed?`repeating-linear-gradient(to right,${C.black} 0px,${C.black} 2px,${C.cream} 2px,${C.cream} 4px)`:C.light,
                  borderBottom:`1px solid ${C.black}`
                }}>
                  <div onClick={()=>removeSrc(src.id)} style={{ width:10,height:10,border:`1px solid ${C.black}`,background:C.cream,cursor:"pointer",fontSize:7,display:"flex",alignItems:"center",justifyContent:"center" }}>✕</div>
                  <span style={{ fontFamily:F,fontSize:9,fontWeight:"bold",background:C.cream,padding:"0 5px" }}>
                    {srcTypes[src.kind].icon} {src.label}
                  </span>
                  <div style={{ width:10 }}/>
                </div>
                <div style={{ padding:5 }}>
                  <FakeScreen kind={src.kind} scene={src.scene} />
                  <div style={{ marginTop:4,display:"flex",flexDirection:"column",gap:3 }}>
                    <Sel value={src.scene} opts={srcTypes[src.kind].scenes} onChange={v=>updateSrc(src.id,{scene:v})} w="100%" />
                    <div style={{ display:"flex",gap:4 }}>
                      <Btn active={src.armed} onClick={()=>updateSrc(src.id,{armed:!src.armed})} style={{ flex:1,fontSize:9,padding:"2px 0" }}>
                        {src.armed?"◉ Armed":"○ Disarmed"}
                      </Btn>
                      <Btn active={src.solo} onClick={()=>updateSrc(src.id,{solo:!src.solo})} style={{ fontSize:9,padding:"2px 6px",color:src.solo?C.cream:C.amber,background:src.solo?C.amber:C.cream }}>
                        S
                      </Btn>
                    </div>
                  </div>
                </div>
              </div>
            ))}

            {/* Add Source */}
            <div style={{
              border:`1.5px dashed ${C.mid}`,borderRadius:3,display:"flex",flexDirection:"column",
              alignItems:"center",justifyContent:"center",minHeight:180,cursor:"pointer",position:"relative"
            }} onClick={()=>setAddOpen(!addOpen)}>
              {!addOpen ? (
                <>
                  <span style={{ fontSize:24,color:C.mid,lineHeight:1 }}>+</span>
                  <span style={{ fontSize:9,color:C.mid,marginTop:4 }}>Add Source</span>
                </>
              ) : (
                <div style={{ display:"flex",flexDirection:"column",gap:4,padding:8 }}>
                  {Object.entries(srcTypes).map(([k,v])=>(
                    <Btn key={k} onClick={e=>{e.stopPropagation();addSrc(k)}} style={{ fontSize:10,textAlign:"left" }}>
                      {v.icon} {k}
                    </Btn>
                  ))}
                </div>
              )}
            </div>
          </div>
        </Win>

        {/* Bottom row: Output + Mic + Status */}
        <div style={{ display:"grid",gridTemplateColumns:"1fr 210px",gap:10 }}>

          {/* Output */}
          <Win title="Output Parameters">
            <div style={{ display:"grid",gridTemplateColumns:"75px 1fr 75px 1fr",gap:"5px 8px",alignItems:"center" }}>
              <span style={{ color:C.mid,fontSize:9 }}>Format:</span>
              <div style={{ display:"flex",gap:8 }}>
                {["MOV","AVI","MP4"].map(f=>(
                  <span key={f} style={{ cursor:"pointer",display:"flex",alignItems:"center",gap:2,fontSize:10 }} onClick={()=>setFormat(f)}>
                    <Radio on={format===f}/>{f}
                  </span>
                ))}
              </div>
              <span style={{ color:C.mid,fontSize:9 }}>Framerate:</span>
              <Sel value={fps} opts={["10 fps","15 fps","24 fps","30 fps"]} onChange={setFps} w={100}/>

              <span style={{ color:C.mid,fontSize:9 }}>Quality:</span>
              <div style={{ display:"flex",alignItems:"center",gap:6 }}>
                <div style={{ flex:1 }}><Slider value={quality} onChange={setQuality}/></div>
                <span style={{ fontSize:9,minWidth:24 }}>{quality}%</span>
              </div>
              <span style={{ color:C.mid,fontSize:9 }}>Audio:</span>
              <Sel value={audio} opts={["22 kHz, 8-bit","44 kHz, 16-bit","48 kHz, 16-bit"]} onChange={setAudio} w={120}/>

              <span style={{ color:C.mid,fontSize:9 }}>Multi-track:</span>
              <div style={{ display:"flex",alignItems:"center",gap:4 }}>
                <Btn active={multiTrack} onClick={()=>setMultiTrack(!multiTrack)} style={{ fontSize:9,padding:"2px 8px" }}>
                  {multiTrack?"◉ Each source → own file":"○ Merge all sources"}
                </Btn>
              </div>
              <span style={{ color:C.mid,fontSize:9 }}>Save to:</span>
              <Sel value="Macintosh HD" opts={["Macintosh HD","Desktop","Documents"]} onChange={()=>{}} w={120}/>
            </div>

            {/* Transport */}
            <div style={{ marginTop:8,paddingTop:7,borderTop:`1px solid ${C.light}`,display:"flex",alignItems:"center",gap:6 }}>
              <Btn accent={recording?undefined:C.red} active={recording}
                onClick={()=>{setRecording(!recording);setPaused(false);if(recording)setElapsed(0)}}
                style={recording?{}:{color:C.cream,background:C.red}}>
                {recording?"◼ Stop":"● Rec"}
              </Btn>
              <Btn onClick={()=>recording&&setPaused(!paused)} active={paused} disabled={!recording}>
                {paused?"▶ Resume":"❚❚ Pause"}
              </Btn>
              <div style={{ flex:1,textAlign:"center" }}>
                {recording && <span style={{ fontSize:8,color:C.mid }}>
                  {multiTrack?`${armed.length} file${armed.length!==1?"s":""}:`:"merged:"} ~{(elapsed*0.4*armed.length).toFixed(1)} MB
                </span>}
              </div>
              <span style={{
                fontFamily:"Monaco,monospace",fontSize:15,fontWeight:"bold",letterSpacing:2,
                color:recording&&!paused?C.red:C.dark,background:"#e0dcd4",padding:"2px 8px",
                borderRadius:2,border:`1px inset ${C.light}`
              }}>{fmt(elapsed)}</span>
            </div>
          </Win>

          {/* Mic + Status stacked */}
          <div style={{ display:"flex",flexDirection:"column",gap:10 }}>
            <Win title="Microphone">
              <div style={{ display:"flex",flexDirection:"column",gap:4 }}>
                <div style={{ display:"flex",alignItems:"center",gap:4 }}><span style={{ fontSize:8,color:C.mid,width:8 }}>L</span><MicMeter level={micLevel}/></div>
                <div style={{ display:"flex",alignItems:"center",gap:4 }}><span style={{ fontSize:8,color:C.mid,width:8 }}>R</span><MicMeter level={micLevel*0.85}/></div>
                <Waveform active={recording&&!paused}/>
                <div style={{ display:"flex",alignItems:"center",gap:4 }}>
                  <span style={{ fontSize:8,color:C.mid,width:26 }}>Input</span>
                  <Sel value={micInput} opts={["Built-in Mic","External","Line In"]} onChange={setMicInput} w={95}/>
                </div>
                <div style={{ display:"flex",alignItems:"center",gap:4 }}>
                  <span style={{ fontSize:8,color:C.mid,width:26 }}>Gain</span>
                  <div style={{ flex:1 }}><Slider value={gain} onChange={setGain}/></div>
                </div>
              </div>
            </Win>

            <Win title="Status">
              <div style={{ display:"flex",flexDirection:"column",gap:4 }}>
                <div style={{ display:"flex",alignItems:"center",gap:4 }}>
                  <span style={{ fontSize:8,color:C.mid }}>Disk</span>
                  <div style={{ flex:1,height:7,background:C.light,border:`1px solid ${C.mid}`,borderRadius:1,overflow:"hidden" }}>
                    <div style={{ width:`${diskPct}%`,height:"100%",background:diskPct>85?C.amber:C.mid,transition:"width 1s" }}/>
                  </div>
                  <span style={{ fontSize:8 }}>{Math.round(100-diskPct)}%</span>
                </div>
                <div style={{ fontSize:9,color:C.mid }}>
                  Status: <span style={{ color:recording?(paused?C.amber:C.red):C.dark,fontWeight:"bold" }}>
                    {recording?(paused?"⏸ Paused":"● Recording"):"◻ Ready"}
                  </span>
                </div>
                <div style={{ fontSize:8,color:C.mid }}>
                  Armed: {armed.map(s=>s.label).join(", ")||"None"}
                </div>
              </div>
            </Win>
          </div>
        </div>
      </div>

      <style>{`@keyframes blink{0%,49%{opacity:1}50%,100%{opacity:0}}`}</style>
    </div>
  );
}
