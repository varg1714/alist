import{f as e,ae as y,ah as d,a7 as I,o,W as l,bf as O,bc as P,ai as h,by as T,bI as $,d as b,t as _,bJ as k,ad as L,a5 as x,B as m,a9 as S,v as A,aa as D,bK as f,m as g,a0 as v,bL as j,z as V,bM as R,G as a,H as i,bG as M,bN as C,e as z,aP as F,P as B,Q as W,ab as H}from"./index.6e7e7f0f.js";import{a as w,b as G}from"./useUtil.c2ec2839.js";import{g as X,O as u}from"./icon.1c0dd91e.js";import{m as U}from"./index.1f69e779.js";import{T as J}from"./Layout.b9d6eada.js";const K=r=>e(l,{class:"fileinfo",py:"$6",spacing:"$6",get children(){return[e(y,{boxSize:"$20",get fallback(){return e(d,{get color(){return I()},boxSize:"$20",get as(){return X(o.obj)}})},get src(){return o.obj.thumb}}),e(l,{spacing:"$2",get children(){return[e(O,{size:"lg",css:{wordBreak:"break-all"},get children(){return o.obj.name}}),e(P,{color:"$neutral10",size:"sm",get children(){return[h(()=>T(o.obj.size))," \xB7 ",h(()=>$(o.obj.modified))]}})]}}),e(l,{spacing:"$2",get children(){return r.children}})]}}),E=()=>{const r=b(),n=_(()=>k(o.obj.name)),{currentObjLink:s}=w();return e(g,{get when(){return n().length},get children(){return e(L,{get children(){return[e(x,{as:m,colorScheme:"success",get rightIcon(){return e(d,{as:U})},get children(){return r("home.preview.open_with")}}),e(S,{get children(){return e(A,{get each(){return n()},children:t=>e(D,{as:"a",target:"_blank",get href(){return f(t.value,{raw_url:o.raw_url,name:o.obj.name,d_url:s(!0)})},get children(){return t.key}})})}})]}})}})},p=r=>{const n=b(),{copyCurrentRawLink:s}=G();return e(K,{get children(){return[e(v,{spacing:"$2",get children(){return[e(m,{colorScheme:"accent",onClick:()=>s(!0),get children(){return n("home.toolbar.copy_link")}}),e(m,{as:"a",get href(){return o.raw_url},target:"_blank",get children(){return n("home.preview.download")}})]}}),e(g,{get when(){return r.openWith},get children(){return e(E,{})}})]}})},N=Object.freeze(Object.defineProperty({__proto__:null,Download:p,default:p},Symbol.toStringTag,{value:"Module"})),Q=r=>{const{currentObjLink:n}=w(),s=_(()=>f(r.scheme,{raw_url:o.raw_url,name:o.obj.name,d_url:n(!0)}));return e(R,{w:"$full",h:"70vh",get children(){return[e(j.iframe,{w:"$full",h:"$full",get src(){return s()}}),e(d,{pos:"absolute",right:"$2",bottom:"$10","aria-label":"Open in new tab",as:J,onClick:()=>{window.open(s(),"_blank")},cursor:"pointer",rounded:"$md",get bgColor(){return V()},p:"$1",boxSize:"$7"})]}})},q=r=>()=>e(Q,{scheme:r}),Y=[{name:"Aliyun Video Previewer",type:u.VIDEO,provider:/^Aliyundrive(Open)?$/,component:a(()=>i(()=>import("./aliyun_video.508a23bc.js"),["assets/aliyun_video.508a23bc.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/icon.1c0dd91e.js","assets/index.d1a6db04.js","assets/index.1f69e779.js","assets/Layout.b9d6eada.js","assets/Markdown.aec033e7.js","assets/index.7aa8ddb3.js","assets/FolderTree.05153b1b.js","assets/hls.8e18d71e.js"]))},{name:"Markdown",type:u.TEXT,component:a(()=>i(()=>import("./markdown.f5a9c258.js"),["assets/markdown.f5a9c258.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/Markdown.aec033e7.js"]))},{name:"Markdown with word wrap",type:u.TEXT,component:a(()=>i(()=>import("./markdown_with_word_wrap.90cb95a2.js"),["assets/markdown_with_word_wrap.90cb95a2.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/Markdown.aec033e7.js"]))},{name:"Text Editor",type:u.TEXT,component:a(()=>i(()=>import("./text-editor.2411a892.js"),["assets/text-editor.2411a892.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js"]))},{name:"HTML render",exts:["html"],component:a(()=>i(()=>import("./html.d5d13fbf.js"),["assets/html.d5d13fbf.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js"]))},{name:"Image",type:u.IMAGE,component:a(()=>i(()=>import("./image.f6473840.js"),["assets/image.f6473840.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/ImageWithError.64e7e53e.js"]))},{name:"Video",type:u.VIDEO,component:a(()=>i(()=>import("./video.47ff1e6c.js"),["assets/video.47ff1e6c.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/icon.1c0dd91e.js","assets/index.d1a6db04.js","assets/index.1f69e779.js","assets/Layout.b9d6eada.js","assets/Markdown.aec033e7.js","assets/index.7aa8ddb3.js","assets/FolderTree.05153b1b.js","assets/hls.8e18d71e.js"]))},{name:"Audio",type:u.AUDIO,component:a(()=>i(()=>import("./audio.85e6bf56.js"),["assets/audio.85e6bf56.js","assets/audio.e5b5af14.css","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/icon.1c0dd91e.js","assets/index.d1a6db04.js","assets/index.1f69e779.js","assets/Layout.b9d6eada.js","assets/Markdown.aec033e7.js","assets/index.7aa8ddb3.js","assets/FolderTree.05153b1b.js"]))},{name:"Ipa",exts:["ipa","tipa"],component:a(()=>i(()=>import("./ipa.a4671e09.js"),["assets/ipa.a4671e09.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/icon.1c0dd91e.js","assets/index.d1a6db04.js","assets/index.1f69e779.js","assets/Layout.b9d6eada.js","assets/Markdown.aec033e7.js","assets/index.7aa8ddb3.js","assets/FolderTree.05153b1b.js"]))},{name:"Plist",exts:["plist"],component:a(()=>i(()=>import("./plist.5624556b.js"),["assets/plist.5624556b.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css","assets/useUtil.c2ec2839.js","assets/api.b90cfaf8.js","assets/icon.1c0dd91e.js","assets/index.d1a6db04.js","assets/index.1f69e779.js","assets/Layout.b9d6eada.js","assets/Markdown.aec033e7.js","assets/index.7aa8ddb3.js","assets/FolderTree.05153b1b.js"]))},{name:"Aliyun Office Previewer",exts:["doc","docx","ppt","pptx","xls","xlsx","pdf"],provider:/^Aliyundrive(Share)?$/,component:a(()=>i(()=>import("./aliyun_office.851fa40d.js"),["assets/aliyun_office.851fa40d.js","assets/index.6e7e7f0f.js","assets/index.15d8d646.css"]))}],Z=r=>{const n=[];return Y.forEach(t=>{var c;t.provider&&!t.provider.test(r.provider)||(t.type===r.type||t.exts==="*"||((c=t.exts)==null?void 0:c.includes(M(r.name).toLowerCase())))&&n.push({name:t.name,component:t.component})}),C(r.name).forEach(t=>{n.push({name:t.key,component:q(t.value)})}),n.push({name:"Download",component:a(()=>i(()=>Promise.resolve().then(()=>N),void 0))}),n},ee=()=>{const r=_(()=>Z({...o.obj,provider:o.provider})),[n,s]=z(r()[0]);return e(g,{get when(){return r().length>1},get fallback(){return e(p,{openWith:!0})},get children(){return e(l,{w:"$full",spacing:"$2",get children(){return[e(v,{w:"$full",spacing:"$2",get children(){return[e(F,{alwaysShowBorder:!0,get value(){return n().name},onChange:t=>{s(r().find(c=>c.name===t))},get options(){return r().map(t=>({value:t.name}))}}),e(E,{})]}}),e(B,{get fallback(){return e(W,{})},get children(){return e(H,{get component(){return n().component}})}})]}})}})},ie=Object.freeze(Object.defineProperty({__proto__:null,default:ee},Symbol.toStringTag,{value:"Module"}));export{K as F,ie as a};
