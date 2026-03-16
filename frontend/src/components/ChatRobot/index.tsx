import { useState } from 'react';
import { ChatgptRobot } from 'react-chatgpt-modal';
import { Button, Tooltip } from 'antd';
import store from '@/store';
import styles from './index.module.css';

import StarIcon from './star.png';
import BotGitIcon from './bot.gif';
import BotIcon from './bot.png';

const ChatRobot = () => {
  const [configModel] = store.useModel('config');
  const configData = configModel.properties || {};
  const [visible, setVisible] = useState(false);

  if (!configData['chat.enabled']) {
    return null;
  }

  const gptApi = configData['chat.endpoint'];

  const initMessage = {
    text: '您好，欢迎使用 <b>Aigateway</b>。',
    img: StarIcon,
    date: new Date(),
    reply: true,
    type: 'init',
    user: {
      name: 'Aigateway',
      avatar: BotIcon,
    },
  };

  const handleClckiBtn = () => {
    setVisible(true);
  };

  return (
    <div className={styles.chatRobot}>
      {!visible && (
        <Tooltip title="GPT Robot" placement="left">
          <Button
            className={styles['chat-robot-btn']}
            size="large"
            type="primary"
            shape="circle"
            onClick={handleClckiBtn}
          >
            <img className={styles['chat-robot-btn-gif']} src={BotGitIcon} />
            <img className={styles['chat-robot-btn-png']} src={BotIcon} />
          </Button>
        </Tooltip>
      )}
      <ChatgptRobot
        onClose={() => setVisible(false)}
        visible={visible}
        config={{
          initMessage,
          title: 'GPT Robot',
          gptApi,
        }}
        robotStyle={{
          right: 10,
          bottom: 10,
        }}
      />
    </div>
  );
};

export default ChatRobot;
