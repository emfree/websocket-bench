extern crate websocket;
extern crate  libc;

use std::iter;
use std::thread;
use websocket::{Server, Message, Sender, Receiver};
use websocket::message::Type;
use websocket::header::WebSocketProtocol;

fn main() {
    let server = Server::bind("0.0.0.0:8001").unwrap();

    for connection in server {
        // Spawn a new thread for each connection.
        thread::Builder::new().stack_size(2 * 1024 * 1024).spawn(move || {
            let request = connection.unwrap().read_request().unwrap(); // Get the request
            let headers = request.headers.clone(); // Keep the headers so we can check them

            request.validate().unwrap(); // Validate the request

            let mut response = request.accept(); // Form a response

            if let Some(&WebSocketProtocol(ref protocols)) = headers.get() {
                if protocols.contains(&("rust-websocket".to_string())) {
                    // We have a protocol we want to use
                    response.headers.set(WebSocketProtocol(vec!["rust-websocket".to_string()]));
                }
            }

            let mut client = response.send().unwrap(); // Send the response

            let ip = client.get_mut_sender()
                .get_mut()
                .peer_addr()
                .unwrap();

            println!("Connection from {}", ip);

            let (mut sender, mut receiver) = client.split();
            sender.get_mut().set_nodelay(true);

            for message in receiver.incoming_messages() {
                let message: Message = message.unwrap();

                match message.opcode {
                    Type::Close => {
                        let message = Message::close();
                        sender.send_message(&message).unwrap();
                        println!("Client {} disconnected", ip);
                        return;
                    }
                    _ => {
                        let cpu_num;
                        unsafe {
                            cpu_num = libc::sched_getcpu();
                        }
                        let mut resp_payload = format!("{}", cpu_num);
                        let padding_len = message.payload.len() - resp_payload.len();
                        let padding: String = iter::repeat(" ").take(padding_len).collect();
                        resp_payload = resp_payload + &padding;

                        let resp = Message::text(resp_payload);
                        sender.send_message(&resp).unwrap()
                    }
                }
            }
        });
    }
}
