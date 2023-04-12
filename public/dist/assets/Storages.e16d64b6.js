import{d as w,u as k,a as f,b6 as l,f as e,J as B,a7 as T,a0 as c,bc as S,bx as L,ai as M,Y as O,B as o,aN as V,bd as g,n as p,W as C,e as $,t as W,m as I,bS as R,bT as A,bU as F,bV as H,bW as N,bX as P,bY as X,v,bZ as Y,b_ as G,b$ as J,c0 as U}from"./index.0596c76d.js";import{o as Z}from"./index.bf378032.js";import{D as j}from"./DeletePopover.a1846e76.js";const q=t=>{const a=w(),{to:i}=k(),[d,u]=f(()=>l.post(`/admin/storage/delete?id=${t.storage.id}`)),[h,s]=f(()=>l.post(`/admin/storage/${t.storage.disabled?"enable":"disable"}?id=${t.storage.id}`));return e(C,{w:"$full",spacing:"$2",rounded:"$lg",border:"1px solid $neutral7",get background(){return B("$neutral2","$neutral3")()},p:"$3",get _hover(){return{border:`1px solid ${T()}`}},get children(){return[e(c,{spacing:"$2",get children(){return[e(S,{fontWeight:"$medium",css:{wordBreak:"break-all"},get children(){return t.storage.mount_path}}),e(L,{colorScheme:"info",get children(){return a(`drivers.drivers.${t.storage.driver}`)}})]}}),e(c,{get children(){return[e(S,{get children(){return[M(()=>a("storages.common.status")),":\xA0"]}}),e(O,{css:{wordBreak:"break-all"},overflowX:"auto",get innerHTML(){return t.storage.status}})]}}),e(S,{css:{wordBreak:"break-all"},get children(){return t.storage.remark}}),e(c,{spacing:"$2",get children(){return[e(o,{onClick:()=>{i(`/@manage/storages/edit/${t.storage.id}`)},get children(){return a("global.edit")}}),e(o,{get loading(){return h()},colorScheme:"warning",onClick:async()=>{const n=await s();V(n,()=>{t.refresh()})},get children(){return a(`global.${t.storage.disabled?"enable":"disable"}`)}}),e(j,{get name(){return t.storage.mount_path},get loading(){return d()},onClick:async()=>{const n=await u();g(n,()=>{p.success(a("global.delete_success")),t.refresh()})}})]}})]}})},ee=()=>{const t=w();Z("manage.sidemenu.storages");const{to:a}=k(),[i,d]=f(()=>l.get("/admin/storage/list")),[u,h]=$([]),s=async()=>{const r=await d();g(r,b=>h(b.content))},[n,_]=$([]),[m,x]=$([]);(async()=>{const r=await l.get("/admin/driver/names");g(r,b=>_(b))})(),s();const D=async()=>{const r=await l.post("/admin/storage/load_all");g(r,()=>{p.success(t("storages.other.start_load_success"))})},y=W(()=>u().filter(r=>m().length===0?!0:m().includes(r.driver)));return e(C,{spacing:"$3",alignItems:"start",w:"$full",get children(){return[e(c,{spacing:"$2",gap:"$2",w:"$full",wrap:{"@initial":"wrap","@md":"unset"},get children(){return[e(o,{colorScheme:"accent",get loading(){return i()},onClick:s,get children(){return t("global.refresh")}}),e(o,{onClick:()=>{a("/@manage/storages/add")},get children(){return t("global.add")}}),e(o,{colorScheme:"warning",get loading(){return i()},onClick:D,get children(){return t("storages.other.load_all")}}),e(I,{get when(){return n().length>0},get children(){return e(R,{multiple:!0,get value(){return m()},onChange:x,get children(){return[e(A,{get children(){return[e(F,{get children(){return t("storages.other.filter_by_driver")}}),e(H,{}),e(N,{})]}}),e(P,{get children(){return e(X,{get children(){return e(v,{get each(){return n()},children:r=>e(Y,{value:r,get children(){return[e(G,{get children(){return t(`drivers.drivers.${r}`)}}),e(J,{})]}})})}})}})]}})}})]}}),e(U,{w:"$full",gap:"$2_5",templateColumns:{"@initial":"1fr","@lg":"repeat(auto-fill, minmax(324px, 1fr))"},get children(){return e(v,{get each(){return y()},children:r=>e(q,{storage:r,refresh:s})})}})]}})};export{ee as default};
