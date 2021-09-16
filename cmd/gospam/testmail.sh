nc -C 127.0.0.1 1025 <<EOF
EHLO localhost
MAIL FROM:<jonas.koeritz@gmail.com>
RCPT TO:<test@localhost>
DATA
Return-Path: <heinz-gustav@post.rwth-aachen.example>
Received: from mx3.gmx.example (qmailr@mx3.gmx.example [195.63.104.129])
  by ancalagon.rhein-neckar.de (8.8.5/8.8.5) with SMTP id SAA25291
  for <karl-heinz@ancalagon.rhein-neckar.de>; Thu, 16 Sep 1998 17:36:20
  +0200 (MET DST)
Received: (qmail 1935 invoked by alias); 16 Sep 1998 15:36:06 -0000
Delivered-To: GMX delivery to karl-heinz@gmx.example
Received: (qmail 27698 invoked by uid 0); 16 Sep 1998 15:36:02 -0000
Received: from pbox.rz.rwth-aachen.example (137.226.144.252)
  by mx3.gmx.example with SMTP; 16 Sep 1998 15:36:02 -0000
Received: from post.rwth-aachen.example (slip-vertech.dialup.RWTH-Aachen.EXAMPLE
  [134.130.73.8]) by pbox.rz.rwth-aachen.example (8.9.1/8.9.0) with ESMTP
  id RAA28830 for <karl-heinz@gmx.example>; Wed, 16 Sep 1998 17:35:59
  +0200
Message-ID: <35FFDA4F.2BC2A064@post.rwth-aachen.example>
Date: Wed, 16 Sep 1998 17:33:35 +0200
From: Heinz-Gustav Hinz <heinz-gustav@post.rwth-aachen.example>
Organization: RWTH Aachen
X-Mailer: Mozilla 4.05 [de] (Win95; I)
To: Karl-Heinz Schmitt <karl-heinz@gmx.example>
MIME-Version: 1.0
Content-Type: text/plain; charset=iso-8859-1
Content-Transfer-Encoding: quoted-printable
Subject: Super long subject that might break the table layoutcompletely lol.
References: <529471993@ancalagon.rhein-neckar.de>
Reply-To: hinz@provider.example
X-Resent-By: Global Message Exchange <forwarder@gmx.example>
X-Resent-For: karl-heinz@gmx.example
X-Resent-To: karl-heinz@ancalagon.rhein-neckar.de

Dies ist eine Testnachricht. Einfach nicht ernst nehmen was in den Headern steht ;).
.
QUIT
EOF