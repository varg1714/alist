import{bE as C,a4 as b,d as E,e as F,a as i,b6 as p,as as G,at as V,f as e,bf as u,m as f,a0 as x,J as W,ah as Z,W as k,bc as h,ai as c,bx as m,bI as J,B as o,r as Q,M as X,i as A,aQ as K,j as U,k as Y,T as ee,I as te,q as re,bd as ne,aN as g}from"./index.6afda90b.js";import{n as ae,u as se}from"./api.db13c611.js";import{G as oe}from"./index.d3a89cd3.js";import ie from"./Common.ee5c328a.js";import"./index.d1f233c0.js";import"./index.73508852.js";import"./SettingItem.d0bfb7d0.js";import"./item_type.be442da4.js";import"./ResponsiveGrid.a357d0a2.js";const de=b('<svg width="1em" height="1em" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-width="2"><path stroke-dasharray="60" stroke-dashoffset="60" stroke-opacity=".3" d="M12 3C16.9706 3 21 7.02944 21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="1.3s" values="60;0"></animate></path><path stroke-dasharray="15" stroke-dashoffset="15" d="M12 3C16.9706 3 21 7.02944 21 12"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.3s" values="15;0"></animate><animateTransform attributeName="transform" dur="1.5s" repeatCount="indefinite" type="rotate" values="0 12 12;360 12 12"></animateTransform></path></g></svg>'),le=b('<svg width="1em" height="1em" viewBox="0 0 24 24"><g stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path fill="currentColor" fill-opacity="0" stroke-dasharray="60" stroke-dashoffset="60" d="M3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12Z"><animate fill="freeze" attributeName="stroke-dashoffset" dur="0.5s" values="60;0"></animate><animate fill="freeze" attributeName="fill-opacity" begin="0.8s" dur="0.15s" values="0;0.3"></animate></path><path fill="none" stroke-dasharray="14" stroke-dashoffset="14" d="M8 12L11 15L16 10"><animate fill="freeze" attributeName="stroke-dashoffset" begin="0.6s" dur="0.2s" values="14;0"></animate></path></g></svg>');function ue(r){return(()=>{const n=de.cloneNode(!0);return C(n,r,!0,!0),n})()}function ce(r){return(()=>{const n=le.cloneNode(!0);return C(n,r,!0,!0),n})()}const ve=()=>{const r=E(),[n,w]=F(),[v,$]=i(()=>p.get("/admin/index/progress")),s=async()=>{const t=await $();ne(t,a=>{w(a)})},y=setInterval(s,5e3);G(()=>clearInterval(y)),s();const[_,I]=i(ae),M=async()=>{const t=await I();g(t),s()},[S,B]=i(()=>p.post("/admin/index/clear")),L=async()=>{const t=await B();g(t),s()},[N,R]=i(()=>p.post("/admin/index/stop")),T=async()=>{const t=await R();g(t),s()};let d,l;const[q,z]=i(se),D=async()=>{let t=[];d.value&&(t=d.value.split(`
`));let a=20;l.value&&(a=parseInt(l.value));const O=await z(t,a);g(O),s()},{isOpen:P,onOpen:j,onClose:H}=V();return e(k,{spacing:"$2",w:"$full",alignItems:"start",get children(){return[e(u,{get children(){return r("manage.sidemenu.settings")}}),e(ie,{get group(){return oe.INDEX}}),e(u,{get children(){return r("indexes.current")}}),e(f,{get when(){return n()},get children(){return e(x,{spacing:"$2",w:"fit-content",shadow:"$md",rounded:"$lg",get bg(){return W("","$neutral3")()},get children(){return[e(Z,{boxSize:"$28",color:"$accent9",p:"$2",get as(){var t;return(t=n())!=null&&t.is_done?ce:ue}}),e(k,{spacing:"$2",flex:"1",alignItems:"start",mr:"$2",get children(){return[e(h,{get children(){return[c(()=>r("indexes.obj_count")),":",e(m,{colorScheme:"info",ml:"$2",get children(){var t;return(t=n())==null?void 0:t.obj_count}})]}}),e(h,{get children(){return[c(()=>r("indexes.last_done_time")),":",e(m,{colorScheme:"accent",ml:"$2",get children(){return c(()=>!!n().last_done_time,!0)()?J(n().last_done_time):r("indexes.unknown")}})]}}),e(f,{get when(){var t;return(t=n())==null?void 0:t.error},get children(){return e(h,{css:{wordBreak:"break-all"},get children(){return[c(()=>r("indexes.error")),":",e(m,{colorScheme:"danger",ml:"$2",get children(){return n().error}})]}})}})]}})]}})}}),e(x,{spacing:"$2",get children(){return[e(o,{colorScheme:"accent",onClick:[s,void 0],get loading(){return v()},get children(){return r("global.refresh")}}),e(o,{colorScheme:"danger",onClick:[L,void 0],get loading(){return S()},get children(){return r("indexes.clear")}}),e(o,{colorScheme:"warning",onClick:[T,void 0],get loading(){return N()},get children(){return r("indexes.stop")}}),e(o,{onClick:[M,void 0],get loading(){return _()},get children(){var t;return r(`indexes.${(t=n())!=null&&t.is_done?"rebuild":"build"}`)}})]}}),e(o,{onClick:j,get children(){return r("indexes.update")}}),e(Q,{get opened(){return P()},onClose:H,get children(){return[e(X,{}),e(A,{get children(){return[e(K,{}),e(U,{get children(){return r("indexes.update")}}),e(Y,{get children(){return[e(u,{get children(){return r("indexes.paths_to_update")}}),e(ee,{ref(t){const a=d;typeof a=="function"?a(t):d=t}}),e(u,{get children(){return r("indexes.max_depth")}}),e(te,{value:20,type:"number",ref(t){const a=l;typeof a=="function"?a(t):l=t}})]}}),e(re,{get children(){return e(o,{onClick:[D,void 0],get loading(){return q()},get children(){return r("indexes.update")}})}})]}})]}})]}})};export{ve as default};
