from fastapi import FastAPI
from pydantic import BaseModel
from openai import OpenAI
from typing import List, Dict, Optional
import json

app = FastAPI()

# 初始化 OpenAI 客户端 ：魔搭提供的免费 api 接口
client = OpenAI(
    base_url='https://api-inference.modelscope.cn/v1/',
    api_key='bc0cd135-9457-43b2-90da-dfd9b30dee19',
)

# 定义消息数据结构
class Sender(BaseModel):
    nickname: str
    userid: int
    card: Optional[str] = None

class Message(BaseModel):
    messagetype: str
    rawmessage: str
    sender: Sender
    time: int

class QQGroupRequest(BaseModel):
    messages: List[Message]

# 幽默标题生成 prompt 模板
HUMOR_PROMPT = """你是一个擅长用幽默方式概括对话的助手。用户将提供一段JSON格式的对话数据，包含对话双方的内容。请执行以下任务：
1. 分析对话的核心主题和笑点
2. 生成一个不超过 20个字的标题
3. 标题要求： 
 - 用谐音梗、双关语或网络热梗 
 - 突出对话中最荒诞/搞笑的部分
 - 避免直白描述（如"关于XX的对话"），但是如果对话中包含🦐/虾姐/饭姐/曹姐/咩/新/公主等要素，可以突出这一元素的存在。

输出格式：
只需返回标题本身，不要包含任何解释、标点或额外文本。

示例：
输入：{"dialogue":[{"role":"A","content":"为什么用微波炉加热葡萄会冒火花？"},{"role":"B","content":"因为葡萄在蹦迪！"}]}
输出：葡星撞地球

现在处理以下JSON对话：
{{user_dialogue}}
"""

def preprocess_qq_messages(messages: List[Message]) -> List[Dict]:
    """
    预处理QQ群消息，过滤掉图片等非文本内容，提取有效对话
    :param messages: QQ群原始消息列表
    :return: 处理后的对话列表
    """
    dialogue = []
    for msg in messages:
        # 跳过图片消息和空消息
        if msg.rawmessage.startswith("[CQ:image") or "已过期" in msg.rawmessage:
            continue
        
        # 添加处理后的消息
        dialogue.append({
            "role": msg.sender.nickname,  # 使用昵称作为角色标识
            "content": msg.rawmessage
        })
    return dialogue

@app.post("/api/qq-humor-title")
async def generate_qq_humor_title(request: QQGroupRequest):
    """
    生成QQ群聊记录的幽默标题
    :param request: 包含QQ群消息的请求体
    :return: 生成的幽默标题
    """
    try:
        # 预处理QQ消息
        dialogue = preprocess_qq_messages(request.messages)
        
        # 构建完整的 prompt
        user_dialogue = {"dialogue": dialogue}
        full_prompt = HUMOR_PROMPT.replace("{{user_dialogue}}", json.dumps(user_dialogue, ensure_ascii=False))
        
        # 调用模型API
        response = client.chat.completions.create(
            model='deepseek-ai/DeepSeek-R1-0528',
            messages=[{
                'role': 'user',
                'content': full_prompt
            }],
            stream=False
        )
        
        # 提取并返回标题
        title = response.choices[0].message.content.strip()
        return {
            "success": True,
            "title": title
        }
    except Exception as e:
        return {
            "success": False,
            "error": str(e)
        }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8888)
