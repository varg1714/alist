import{d as v,u as $,a as l,b6 as o,bw as w,bP as C,e as I,f as t,bQ as R,ap as L,bd as r,n as p,a0 as T,B as u,W as _}from"./index.25b8e5d6.js";import{o as x}from"./index.ddf16d79.js";import{I as B}from"./SettingItem.06b1c70b.js";import{R as P}from"./ResponsiveGrid.805fcc15.js";import"./index.0ad5594c.js";import"./index.afcbe07f.js";import"./item_type.be442da4.js";const j=d=>{const s=v(),{pathname:m}=$();x(`manage.sidemenu.${m().split("/").pop()}`);const[h,f]=l(()=>o.get(`/admin/setting/list?group=${d.group}`)),[i,c]=w([]),a=async()=>{const e=await f();r(e,c)};a();const[b,S]=l(()=>o.post("/admin/setting/save",C(i))),[k,g]=I(!1);return t(_,{w:"$full",alignItems:"start",spacing:"$2",get children(){return[t(P,{get children(){return t(R,{each:i,children:(e,D)=>t(B,L(e,{onChange:n=>{c(y=>e().key===y.key,"value",n)},onDelete:async()=>{g(!0);const n=await o.post(`/admin/setting/delete?key=${e().key}`);g(!1),r(n,()=>{p.success(s("global.delete_success")),a()})}}))})}}),t(T,{spacing:"$2",get children(){return[t(u,{colorScheme:"accent",onClick:a,get loading(){return h()||k()},get children(){return s("global.refresh")}}),t(u,{get loading(){return b()},onClick:async()=>{const e=await S();r(e,()=>p.success(s("global.save_success")))},get children(){return s("global.save")}})]}})]}})};export{j as default};
