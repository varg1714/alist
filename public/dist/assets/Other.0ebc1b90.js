import{d as F,e as i,a as o,b6 as c,f as t,bf as m,c5 as A,ap as g,B as u,bd as l,n as h,I as N,a0 as P,Z}from"./index.c7be97ab.js";import{o as j}from"./index.f3072a2b.js";import{c as z}from"./useUtil.a37c2ef7.js";import{G}from"./index.2247abfb.js";import{I as d}from"./SettingItem.d95d62e8.js";import"./api.2e4c6f66.js";import"./index.abc04cae.js";import"./item_type.be442da4.js";const re=()=>{const r=F();j("manage.sidemenu.other");const[p,_]=i(""),[k,y]=i(""),[f,b]=i(""),[$,v]=i(""),[S,q]=i(""),[a,Q]=i([]),[U,B]=o(()=>c.get(`/admin/setting/list?groups=${G.ARIA2},${G.SINGLE}`)),[H,M]=o(()=>c.post("/admin/setting/set_aria2",{uri:p(),secret:k()})),[O,R]=o(()=>c.post("/admin/setting/set_qbit",{url:f(),seedtime:$()}));(async()=>{const e=await B();l(e,n=>{var C,T,I,L,w;_(((C=n.find(s=>s.key==="aria2_uri"))==null?void 0:C.value)||""),y(((T=n.find(s=>s.key==="aria2_secret"))==null?void 0:T.value)||""),q(((I=n.find(s=>s.key==="token"))==null?void 0:I.value)||""),b(((L=n.find(s=>s.key==="qbittorrent_url"))==null?void 0:L.value)||""),v(((w=n.find(s=>s.key==="qbittorrent_seedtime"))==null?void 0:w.value)||""),Q(n)})})();const[x,D]=o(()=>c.post("/admin/setting/reset_token")),{copy:E}=z();return t(Z,{get loading(){return U()},get children(){return[t(m,{mb:"$2",get children(){return r("settings_other.aria2")}}),t(A,{gap:"$2",columns:{"@initial":1,"@md":2},get children(){return[t(d,g(()=>a().find(e=>e.key==="aria2_uri"),{get value(){return p()},onChange:e=>_(e)})),t(d,g(()=>a().find(e=>e.key==="aria2_secret"),{get value(){return k()},onChange:e=>y(e)}))]}}),t(u,{my:"$2",get loading(){return H()},onClick:async()=>{const e=await M();l(e,n=>{h.success(`${r("settings_other.aria2_version")} ${n}`)})},get children(){return r("settings_other.set_aria2")}}),t(m,{my:"$2",get children(){return r("settings_other.qbittorrent")}}),t(A,{gap:"$2",columns:{"@initial":1,"@md":2},get children(){return[t(d,g(()=>a().find(e=>e.key==="qbittorrent_url"),{get value(){return f()},onChange:e=>b(e)})),t(d,g(()=>a().find(e=>e.key==="qbittorrent_seedtime"),{get value(){return $()},onChange:e=>v(e)}))]}}),t(u,{my:"$2",get loading(){return O()},onClick:async()=>{const e=await R();l(e,n=>{h.success(n)})},get children(){return r("settings_other.set_qbit")}}),t(m,{my:"$2",get children(){return r("settings.token")}}),t(N,{get value(){return S()},readOnly:!0}),t(P,{my:"$2",spacing:"$2",get children(){return[t(u,{onClick:()=>{E(S())},get children(){return r("settings_other.copy_token")}}),t(u,{colorScheme:"danger",get loading(){return x()},onClick:async()=>{const e=await D();l(e,n=>{h.success(r("settings_other.reset_token_success")),q(n)})},get children(){return r("settings_other.reset_token")}})]}})]}})};export{re as default};
